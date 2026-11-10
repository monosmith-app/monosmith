package dev.monosmith.jetbrains.readcommands

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_READCOMMANDS_MONOSMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_READCOMMANDS_MONOSMITH_WHY_WOULD_A_TASK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_READCOMMANDS_MONOSMITH_WHY_WOULD_A_TASK_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_READCOMMANDS_SMITH_COULD_NOT_SHOW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_READCOMMANDS_THE_PROJECT_AND_THE_TASK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_READCOMMANDS_WHY_WOULD_RUN_AGAIN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_READCOMMANDS_WRITE_PROJECT_TASK_FOR_EXAMPLE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
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
    ) = BackgroundFlowClient.start(
        ide,
        say(JETBRAINS_READCOMMANDS_MONOSMITH, what.toString()),
    ) { run(IdePortsClient.of(ide, root), what, args) }

    /** Starts "why would a task run again?", which asks for the project and the task first. */
    fun startWhyMiss(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_READCOMMANDS_MONOSMITH_WHY_WOULD_A_TASK)) { whyMiss(IdePortsClient.of(ide, root)) }

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
            ports.notice.warn(
                say(JETBRAINS_READCOMMANDS_SMITH_COULD_NOT_SHOW, what.toString(), FailureTextMapper.describe(done).toString()),
            )
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
                say(JETBRAINS_READCOMMANDS_MONOSMITH_WHY_WOULD_A_TASK_2),
                say(JETBRAINS_READCOMMANDS_THE_PROJECT_AND_THE_TASK),
            ) { if (taskName.matches(it)) null else say(JETBRAINS_READCOMMANDS_WRITE_PROJECT_TASK_FOR_EXAMPLE) } ?: return false
        return run(ports, say(JETBRAINS_READCOMMANDS_WHY_WOULD_RUN_AGAIN, target.toString()), listOf("why-miss", target))
    }
}
