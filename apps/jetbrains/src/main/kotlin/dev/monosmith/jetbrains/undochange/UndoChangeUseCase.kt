package dev.monosmith.jetbrains.undochange

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
import java.nio.file.Path

/**
 * Reverts the last operation smith committed (`smith undo`). It runs `--dry-run`, shows what would be
 * reverted, and only on a yes reverts with `--yes`. smith itself refuses when a later commit or an
 * uncommitted change touches the same files, and says why.
 */
object UndoChangeUseCase {
    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: undo the last change") {
        val plan = SmithRunClient.run(ide, root, listOf("undo", "--dry-run"))
        if (!plan.ok) return@start NoticeClient.warn(ide, "Nothing was undone. ${FailureTextMapper.describe(plan)}")
        PromptClient.plan(
            ide,
            "MonoSmith: undo the last change?",
            "smith reverts the last change it committed with git revert: a new commit, and your history stays.",
            plan.stdout.trim(),
            "Undo",
            offerCommit = false,
        ) ?: return@start
        val done = SmithRunClient.run(ide, root, listOf("undo", "--yes"), show = true)
        WorkspaceChangedStore.changed(ide)
        if (done.ok) {
            NoticeClient.info(
                ide,
                done.lines.lastOrNull() ?: "The last change is undone.",
            )
        } else {
            NoticeClient.warn(ide, "Nothing was undone. ${FailureTextMapper.describe(done)}")
        }
    }
}
