package dev.monosmith.jetbrains.undochange

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: undo the last change") { run(IdePortsClient.of(ide, root)) }

    /**
     * Runs the flow with the given ports.
     *
     * @return true when smith reverted the last change
     */
    fun run(ports: FlowPortsContract): Boolean {
        val plan = ports.smith.run(listOf("undo", "--dry-run"))
        if (!plan.ok) {
            ports.notice.warn("Nothing was undone. ${FailureTextMapper.describe(plan)}")
            return false
        }
        ports.prompt.plan(
            "MonoSmith: undo the last change?",
            "smith reverts the last change it committed with git revert: a new commit, and your history stays.",
            plan.stdout.trim(),
            "Undo",
            offerCommit = false,
        ) ?: return false
        val done = ports.smith.run(listOf("undo", "--yes"), show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn("Nothing was undone. ${FailureTextMapper.describe(done)}")
            return false
        }
        ports.notice.info(done.lines.lastOrNull() ?: "The last change is undone.")
        return true
    }
}
