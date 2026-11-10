package dev.monosmith.jetbrains.runacross

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_DID_NOT_FINISH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_IS_DONE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_MONOSMITH_RUN_AFFECTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_MONOSMITH_RUN_IN_ALL_PROJECTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_NOTHING_RAN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_NOTHING_RAN_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_NO_PROJECT_HAS_A_TASK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_NO_PROJECT_IS_AFFECTED_BY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_PICK_THE_TASK_TO_RUN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_PICK_THE_TASK_TO_RUN_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RUNACROSS_SMITH_PRINTED_THE_AFFECTED_PROJECTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import dev.monosmith.jetbrains.workspace.AffectedJsonMapper
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.SmithProjectContract
import java.nio.file.Path

/** Runs one task in every project, or in the projects the changes affect: the task is picked from what the projects have. */
object RunAcrossUseCase {
    /** Starts the flow in a background task; [affected] runs only the affected projects. */
    fun start(
        ide: Project,
        root: Path,
        affected: Boolean,
    ) = BackgroundFlowClient.start(
        ide,
        if (affected) say(JETBRAINS_RUNACROSS_MONOSMITH_RUN_AFFECTED) else say(JETBRAINS_RUNACROSS_MONOSMITH_RUN_IN_ALL_PROJECTS),
    ) {
        val projects =
            ListProjectsUseCase.list(root).getOrElse {
                return@start IdePortsClient.of(ide, root).notice.warn(say(JETBRAINS_RUNACROSS_NOTHING_RAN, it.message.toString()))
            }
        run(IdePortsClient.of(ide, root), projects, affected)
    }

    /**
     * Picks the task and runs it.
     *
     * @return true when smith ran the task (whatever the task's own result)
     */
    fun run(
        ports: FlowPortsContract,
        projects: List<SmithProjectContract>,
        affected: Boolean,
    ): Boolean {
        val names = if (affected) readAffected(ports) ?: return false else emptySet()
        val rows = RunAcrossAlgorithm.taskRows(projects, names, affected)
        if (rows.isEmpty()) {
            ports.notice.info(
                if (affected) say(JETBRAINS_RUNACROSS_NO_PROJECT_IS_AFFECTED_BY) else say(JETBRAINS_RUNACROSS_NO_PROJECT_HAS_A_TASK),
            )
            return false
        }
        val title =
            if (affected) say(JETBRAINS_RUNACROSS_MONOSMITH_RUN_AFFECTED) else say(JETBRAINS_RUNACROSS_MONOSMITH_RUN_IN_ALL_PROJECTS)
        val header =
            if (affected) {
                say(JETBRAINS_RUNACROSS_PICK_THE_TASK_TO_RUN)
            } else {
                say(JETBRAINS_RUNACROSS_PICK_THE_TASK_TO_RUN_2)
            }
        val task = ports.prompt.pickOne(title, header, rows, "Run")?.name ?: return false
        val args = RunAcrossAlgorithm.arguments(task, affected) ?: return false
        val done = ports.smith.run(args, show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(
                say(JETBRAINS_RUNACROSS_DID_NOT_FINISH, task.toString(), FailureTextMapper.describe(done).toString()),
            )
        } else {
            ports.notice.info(say(JETBRAINS_RUNACROSS_IS_DONE, task.toString()))
        }
        return true
    }

    private fun readAffected(ports: FlowPortsContract): Set<String>? {
        val run = ports.smith.run(listOf("affected", "--json"))
        if (!run.ok) {
            ports.notice.warn(say(JETBRAINS_RUNACROSS_NOTHING_RAN_2, FailureTextMapper.describe(run).toString()))
            return null
        }
        return runCatching { AffectedJsonMapper.toAffected(run.stdout).projects }.getOrElse {
            ports.notice.warn(say(JETBRAINS_RUNACROSS_SMITH_PRINTED_THE_AFFECTED_PROJECTS))
            null
        }
    }
}
