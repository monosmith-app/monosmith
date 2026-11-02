package dev.monosmith.jetbrains.installtools

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunContract
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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: install the tools") { run(IdePortsClient.of(ide, root)) }

    /**
     * Runs the flow with the given ports.
     *
     * @return true when the person was asked about at least one tool
     */
    fun run(ports: FlowPortsContract): Boolean {
        val planned = ports.smith.run(listOf("tools", "install", "--dry-run", "--json"))
        if (!planned.ok) {
            ports.notice.warn("Nothing was installed. ${FailureTextMapper.describe(planned)}")
            return false
        }
        val plans =
            runCatching { ToolPlansJsonMapper.toPlans(planned.stdout) }
                .getOrElse {
                    ports.notice.warn("smith tools install printed something that is not a plan.")
                    return false
                }
        if (plans.isEmpty()) {
            ports.notice.info("Every tool the workspace needs is installed.")
            return false
        }
        val told = mutableListOf<String>()
        for (plan in plans) {
            val header = "${plan.title} is ${plan.state}. ${plan.why} Without it: ${plan.consequence}. The first answer is the default."
            val way = ports.prompt.pickOne("MonoSmith: get ${plan.title}", header, ToolRowsAlgorithm.wayRows(plan), "Do it") ?: break
            told.add(carryOut(ports, plan, plan.options.first { it.value == way.name }))
        }
        ports.changed()
        if (told.isNotEmpty()) ports.notice.info(told.joinToString("\n"))
        return told.isNotEmpty()
    }

    private fun carryOut(
        ports: FlowPortsContract,
        plan: ToolPlanContract,
        option: ToolOptionContract,
    ): String =
        when (option.kind) {
            "page" -> {
                ports.openUrl(plan.download)
                "Opened ${plan.download}. Install ${plan.title}, then restart the IDE terminal."
            }

            "myself" -> {
                ports.prompt.plan(
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
                install(ports, plan, option)
            }
        }

    private fun install(
        ports: FlowPortsContract,
        plan: ToolPlanContract,
        option: ToolOptionContract,
    ): String {
        val args = listOf("tools", "install", plan.name, "--manager", option.manager)
        val command = option.command.joinToString(" ")
        if (option.admin) {
            val header = "$command needs administrator rights. smith never runs it without this second question."
            val answer =
                ports.prompt.pickOne(
                    "MonoSmith: ${plan.title} needs administrator rights",
                    header,
                    ToolRowsAlgorithm.adminRows(plan, option),
                    "Do it",
                )
            return when (answer?.name) {
                ToolRowsAlgorithm.RUN_IT -> report(plan, ports.smith.run(args + "--with-admin", show = true))
                ToolRowsAlgorithm.MYSELF -> "Run this yourself, then restart the IDE terminal: $command"
                else -> "Skipped ${plan.title}: ${plan.consequence}."
            }
        }
        return report(plan, ports.smith.run(args, show = true))
    }

    private fun report(
        plan: ToolPlanContract,
        run: SmithRunContract,
    ): String =
        if (run.ok) "${plan.title} is installed." else "The install of ${plan.title} did not finish. ${FailureTextMapper.describe(run)}"
}
