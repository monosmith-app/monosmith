package dev.monosmith.jetbrains.runacross

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
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
    ) = BackgroundFlowClient.start(ide, if (affected) "MonoSmith: run affected" else "MonoSmith: run in all projects") {
        val projects =
            ListProjectsUseCase.list(root).getOrElse {
                return@start IdePortsClient.of(ide, root).notice.warn("Nothing ran. ${it.message}")
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
            ports.notice.info(if (affected) "No project is affected by the changes." else "No project has a task.")
            return false
        }
        val title = if (affected) "MonoSmith: run affected" else "MonoSmith: run in all projects"
        val header =
            if (affected) {
                "Pick the task to run in the affected projects. The first answer is the default."
            } else {
                "Pick the task to run in every project that has it."
            }
        val task = ports.prompt.pickOne(title, header, rows, "Run")?.name ?: return false
        val args = RunAcrossAlgorithm.arguments(task, affected) ?: return false
        val done = ports.smith.run(args, show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(
                "$task did not finish. ${FailureTextMapper.describe(done)}",
            )
        } else {
            ports.notice.info("$task is done.")
        }
        return true
    }

    private fun readAffected(ports: FlowPortsContract): Set<String>? {
        val run = ports.smith.run(listOf("affected", "--json"))
        if (!run.ok) {
            ports.notice.warn("Nothing ran. ${FailureTextMapper.describe(run)}")
            return null
        }
        return runCatching { AffectedJsonMapper.toAffected(run.stdout).projects }.getOrElse {
            ports.notice.warn("smith printed the affected projects in a form this plugin does not read. Update the plugin.")
            null
        }
    }
}
