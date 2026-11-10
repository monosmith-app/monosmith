package dev.monosmith.jetbrains.installtools

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.ADOPT_FOUND_CLOSE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DERIVEFILES_DO_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_INSTALL_YOURSELF_THEN_RESTART_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_IS_INSTALLED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_IS_WITHOUT_IT_THE_FIRST
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_MONOSMITH_GET
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_MONOSMITH_INSTALL_THE_TOOLS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_MONOSMITH_INSTALL_YOURSELF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_MONOSMITH_NEEDS_ADMINISTRATOR_RIGHTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_NEEDS_ADMINISTRATOR_RIGHTS_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_NOTHING_WAS_INSTALLED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_OPENED_INSTALL_THEN_RESTART_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_RUN_ONE_OF_THESE_THEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_RUN_THIS_YOURSELF_THEN_RESTART
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_SKIPPED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_SMITH_TOOLS_INSTALL_PRINTED_SOMETHING
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_INSTALLTOOLS_THE_INSTALL_OF_DID_NOT
import dev.monosmith.jetbrains.words.SmithWordsConfig.TOOLSDIALOG_DONE_NONE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_INSTALLTOOLS_MONOSMITH_INSTALL_THE_TOOLS)) { run(IdePortsClient.of(ide, root)) }

    /**
     * Runs the flow with the given ports.
     *
     * @return true when the person was asked about at least one tool
     */
    fun run(ports: FlowPortsContract): Boolean {
        val planned = ports.smith.run(listOf("tools", "install", "--dry-run", "--json"))
        if (!planned.ok) {
            ports.notice.warn(say(JETBRAINS_INSTALLTOOLS_NOTHING_WAS_INSTALLED, FailureTextMapper.describe(planned).toString()))
            return false
        }
        val plans =
            runCatching { ToolPlansJsonMapper.toPlans(planned.stdout) }
                .getOrElse {
                    ports.notice.warn(say(JETBRAINS_INSTALLTOOLS_SMITH_TOOLS_INSTALL_PRINTED_SOMETHING))
                    return false
                }
        if (plans.isEmpty()) {
            ports.notice.info(say(TOOLSDIALOG_DONE_NONE))
            return false
        }
        val told = mutableListOf<String>()
        for (plan in plans) {
            val header =
                say(
                    JETBRAINS_INSTALLTOOLS_IS_WITHOUT_IT_THE_FIRST,
                    plan.title.toString(),
                    plan.state.toString(),
                    plan.why.toString(),
                    plan.consequence.toString(),
                )
            val way =
                ports.prompt.pickOne(
                    say(JETBRAINS_INSTALLTOOLS_MONOSMITH_GET, plan.title.toString()),
                    header,
                    ToolRowsAlgorithm.wayRows(plan),
                    say(JETBRAINS_DERIVEFILES_DO_IT),
                ) ?: break
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
                say(JETBRAINS_INSTALLTOOLS_OPENED_INSTALL_THEN_RESTART_THE, plan.download.toString(), plan.title.toString())
            }

            "myself" -> {
                ports.prompt.plan(
                    say(JETBRAINS_INSTALLTOOLS_MONOSMITH_INSTALL_YOURSELF, plan.title.toString()),
                    say(JETBRAINS_INSTALLTOOLS_RUN_ONE_OF_THESE_THEN),
                    ToolRowsAlgorithm.commands(plan).joinToString("\n"),
                    say(ADOPT_FOUND_CLOSE),
                    offerCommit = false,
                )
                say(JETBRAINS_INSTALLTOOLS_INSTALL_YOURSELF_THEN_RESTART_THE, plan.title.toString())
            }

            "skip" -> {
                say(JETBRAINS_INSTALLTOOLS_SKIPPED, plan.title.toString(), plan.consequence.toString())
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
            val header = say(JETBRAINS_INSTALLTOOLS_NEEDS_ADMINISTRATOR_RIGHTS_SMITH, command.toString())
            val answer =
                ports.prompt.pickOne(
                    say(JETBRAINS_INSTALLTOOLS_MONOSMITH_NEEDS_ADMINISTRATOR_RIGHTS, plan.title.toString()),
                    header,
                    ToolRowsAlgorithm.adminRows(plan, option),
                    say(JETBRAINS_DERIVEFILES_DO_IT),
                )
            return when (answer?.name) {
                ToolRowsAlgorithm.RUN_IT -> report(plan, ports.smith.run(args + "--with-admin", show = true))
                ToolRowsAlgorithm.MYSELF -> say(JETBRAINS_INSTALLTOOLS_RUN_THIS_YOURSELF_THEN_RESTART, command.toString())
                else -> say(JETBRAINS_INSTALLTOOLS_SKIPPED, plan.title.toString(), plan.consequence.toString())
            }
        }
        return report(plan, ports.smith.run(args, show = true))
    }

    private fun report(
        plan: ToolPlanContract,
        run: SmithRunContract,
    ): String =
        if (run.ok) {
            say(
                JETBRAINS_INSTALLTOOLS_IS_INSTALLED,
                plan.title.toString(),
            )
        } else {
            say(JETBRAINS_INSTALLTOOLS_THE_INSTALL_OF_DID_NOT, plan.title.toString(), FailureTextMapper.describe(run).toString())
        }
}
