package dev.monosmith.jetbrains.readcommands

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Commands of smith that only read: the cache status and its check, the tools the workspace needs, the
 * installed smith with its settings, the copies between projects, the Docker Hub repositories a description
 * would reach, and why a task would run again. They change nothing, so they ask nothing but a task name;
 * smith prints in the Output tab.
 */
object ReadSmithCommandUseCase {
    private val taskName = Regex("^[^\\s:]+:[^\\s:]+$")

    /** Starts the command in a background task; [what] names what it shows, and [args] is the command. */
    fun start(
        ide: Project,
        root: Path,
        what: String,
        args: List<String>,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: $what") { run(IdePortsClient.of(ide, root), what, args) }

    /** Starts "why would a task run again?", which asks for the project and the task first. */
    fun startWhyMiss(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: why would a task run again") { whyMiss(IdePortsClient.of(ide, root)) }

    /**
     * Runs a command that reads and shows its output in the Output tab.
     *
     * @param what what the command shows, for the notice when it fails
     * @param args the command, after `smith`
     * @return true when smith ended with code 0
     */
    fun run(
        ports: FlowPortsContract,
        what: String,
        args: List<String>,
    ): Boolean {
        val done = ports.smith.run(args, show = true)
        if (!done.ok) {
            ports.notice.warn("smith could not show $what. ${FailureTextMapper.describe(done)}")
            return false
        }
        return true
    }

    /**
     * Asks for `project:task` and shows why smith would run that task again (`smith why-miss`).
     *
     * @return true when smith printed the answer
     */
    fun whyMiss(ports: FlowPortsContract): Boolean {
        val target =
            ports.prompt.text(
                "MonoSmith: why would a task run again?",
                "The project and the task, as project:task (for example web:build):",
            ) { if (taskName.matches(it)) null else "Write project:task, for example web:build." } ?: return false
        return run(ports, "why $target would run again", listOf("why-miss", target))
    }
}
