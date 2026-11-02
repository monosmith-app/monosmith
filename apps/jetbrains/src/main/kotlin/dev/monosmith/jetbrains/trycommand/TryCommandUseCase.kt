package dev.monosmith.jetbrains.trycommand

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Runs a smith command that changes files in a temporary copy (`smith try <command> --json`), shows what
 * it did and which checks passed, and, only after a try that passed and a yes, runs it for real
 * (`smith try <command> --yes`). The try itself leaves the files of the person as they are.
 */
object TryCommandUseCase {
    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: try a command") { run(IdePortsClient.of(ide, root)) }

    /**
     * Runs the flow with the given ports.
     *
     * @return true when the command ran for real
     */
    fun run(ports: FlowPortsContract): Boolean {
        val typed =
            ports.prompt.text(
                "MonoSmith: try a command in a copy",
                "The smith command to try, the words after smith (for example add lib utils --lang node):",
            ) {
                if (it.isEmpty()) "Type the command to try." else null
            } ?: return false
        val command = typed.removePrefix("smith ").split(Regex("\\s+")).filter { it.isNotEmpty() }
        val tried = ports.smith.run(listOf("try") + command + "--json")
        val shown =
            runCatching {
                TryResultMapper.toPage(tried.stdout)
            }.getOrElse {
                ports.notice.warn("Nothing was tried. ${FailureTextMapper.describe(tried)}")
                return false
            }
        val answer =
            ports.prompt.plan(
                "MonoSmith: try smith ${command.joinToString(" ")}",
                if (shown.passed) {
                    "It passed in a copy. Run it for real? Nothing is committed."
                } else {
                    "It failed in the copy, so it is not offered for real."
                },
                shown.page,
                if (shown.passed) "Run it for real" else "Close",
                offerCommit = false,
            )
        if (answer == null || !shown.passed) return false
        val done = ports.smith.run(listOf("try") + command + "--yes", show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn("Nothing was changed. ${FailureTextMapper.describe(done)}")
            return false
        }
        ports.notice.info("The command ran for real.")
        return true
    }
}
