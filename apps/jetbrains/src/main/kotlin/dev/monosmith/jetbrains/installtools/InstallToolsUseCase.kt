package dev.monosmith.jetbrains.installtools

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
import java.nio.file.Path

/**
 * Offers the ways to get each tool that is missing, and carries out the answers. smith prints the plan
 * (`tools install --dry-run --json`), the person picks for each tool with the command in front of them,
 * and smith runs only that command. A command that needs administrator rights is asked about again.
 */
object InstallToolsUseCase {
    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: install the tools") {
        val planned = SmithRunClient.run(ide, root, listOf("tools", "install", "--dry-run", "--json"))
        if (!planned.ok) return@start NoticeClient.warn(ide, "Nothing was installed. ${FailureTextMapper.describe(planned)}")
        val plans =
            runCatching { ToolPlansJsonMapper.toPlans(planned.stdout) }
                .getOrElse { return@start NoticeClient.warn(ide, "smith tools install printed something that is not a plan.") }
        if (plans.isEmpty()) return@start NoticeClient.info(ide, "Every tool the workspace needs is installed.")
        val told = mutableListOf<String>()
        for (plan in plans) {
            val header = "${plan.title} is ${plan.state}. ${plan.why} Without it: ${plan.consequence}. The first answer is the default."
            val way = PromptClient.pickOne(ide, "MonoSmith: get ${plan.title}", header, ToolRowsAlgorithm.wayRows(plan), "Do it") ?: break
            told.add(carryOut(ide, root, plan, plan.options.first { it.value == way.name }))
        }
        WorkspaceChangedStore.changed(ide)
        if (told.isNotEmpty()) NoticeClient.info(ide, told.joinToString("\n"))
    }

    private fun carryOut(
        ide: Project,
        root: Path,
        plan: ToolPlanContract,
        option: ToolOptionContract,
    ): String =
        when (option.kind) {
            "page" -> {
                PromptClient.onUi { BrowserUtil.browse(plan.download) }
                "Opened ${plan.download}. Install ${plan.title}, then restart the IDE terminal."
            }

            "myself" -> {
                PromptClient.plan(
                    ide,
                    "MonoSmith: install ${plan.title} yourself",
                    "Run one of these, then restart the IDE terminal.",
                    ToolRowsAlgorithm.commands(plan).joinToString("\n"),
                    "Close",
                    offerCommit = false,
                )
                "Install ${plan.title} yourself, then restart the IDE terminal."
            }

            "skip" -> {
                "Skipped ${plan.title}: ${plan.consequence}."
            }

            else -> {
                install(ide, root, plan, option)
            }
        }

    private fun install(
        ide: Project,
        root: Path,
        plan: ToolPlanContract,
        option: ToolOptionContract,
    ): String {
        val args = listOf("tools", "install", plan.name, "--manager", option.manager)
        val command = option.command.joinToString(" ")
        if (option.admin) {
            val header = "$command needs administrator rights. smith never runs it without this second question."
            val answer =
                PromptClient.pickOne(
                    ide,
                    "MonoSmith: ${plan.title} needs administrator rights",
                    header,
                    ToolRowsAlgorithm.adminRows(plan, option),
                    "Do it",
                )
            return when (answer?.name) {
                ToolRowsAlgorithm.RUN_IT -> report(plan, SmithRunClient.run(ide, root, args + "--with-admin", show = true))
                ToolRowsAlgorithm.MYSELF -> "Run this yourself, then restart the IDE terminal: $command"
                else -> "Skipped ${plan.title}: ${plan.consequence}."
            }
        }
        return report(plan, SmithRunClient.run(ide, root, args, show = true))
    }

    private fun report(
        plan: ToolPlanContract,
        run: dev.monosmith.jetbrains.smithrun.SmithRunContract,
    ): String =
        if (run.ok) "${plan.title} is installed." else "The install of ${plan.title} did not finish. ${FailureTextMapper.describe(run)}"
}
