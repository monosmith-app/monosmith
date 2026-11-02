package dev.monosmith.jetbrains.trycommand

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: try a command") {
        val typed =
            PromptClient.text(
                ide,
                "MonoSmith: try a command in a copy",
                "The smith command to try, the words after smith (for example add lib utils --lang node):",
            ) {
                if (it.isEmpty()) "Type the command to try." else null
            } ?: return@start
        val command = typed.removePrefix("smith ").split(Regex("\\s+")).filter { it.isNotEmpty() }
        val tried = SmithRunClient.run(ide, root, listOf("try") + command + "--json")
        val shown =
            runCatching {
                TryResultMapper.toPage(tried.stdout)
            }.getOrElse { return@start NoticeClient.warn(ide, "Nothing was tried. ${FailureTextMapper.describe(tried)}") }
        val answer =
            PromptClient.plan(
                ide,
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
        if (answer == null || !shown.passed) return@start
        val done = SmithRunClient.run(ide, root, listOf("try") + command + "--yes", show = true)
        WorkspaceChangedStore.changed(ide)
        if (done.ok) {
            NoticeClient.info(
                ide,
                "The command ran for real.",
            )
        } else {
            NoticeClient.warn(ide, "Nothing was changed. ${FailureTextMapper.describe(done)}")
        }
    }
}
