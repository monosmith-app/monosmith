package dev.monosmith.jetbrains.undochange

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_UNDO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_UNDOCHANGE_MONOSMITH_UNDO_THE_LAST_CHANGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_UNDOCHANGE_MONOSMITH_UNDO_THE_LAST_CHANGE_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_UNDOCHANGE_NOTHING_WAS_UNDONE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_UNDOCHANGE_SMITH_REVERTS_THE_LAST_CHANGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_UNDOCHANGE_THE_LAST_CHANGE_IS_UNDONE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_UNDOCHANGE_MONOSMITH_UNDO_THE_LAST_CHANGE)) { run(IdePortsClient.of(ide, root)) }

    /**
     * Runs the flow with the given ports.
     *
     * @return true when smith reverted the last change
     */
    fun run(ports: FlowPortsContract): Boolean {
        val plan = ports.smith.run(listOf("undo", "--dry-run"))
        if (!plan.ok) {
            ports.notice.warn(say(JETBRAINS_UNDOCHANGE_NOTHING_WAS_UNDONE, FailureTextMapper.describe(plan).toString()))
            return false
        }
        ports.prompt.plan(
            say(JETBRAINS_UNDOCHANGE_MONOSMITH_UNDO_THE_LAST_CHANGE_2),
            say(JETBRAINS_UNDOCHANGE_SMITH_REVERTS_THE_LAST_CHANGE),
            plan.stdout.trim(),
            say(JETBRAINS_ADDPROJECT_UNDO),
            offerCommit = false,
        ) ?: return false
        val done = ports.smith.run(listOf("undo", "--yes"), show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(say(JETBRAINS_UNDOCHANGE_NOTHING_WAS_UNDONE, FailureTextMapper.describe(done).toString()))
            return false
        }
        ports.notice.info(done.lines.lastOrNull() ?: say(JETBRAINS_UNDOCHANGE_THE_LAST_CHANGE_IS_UNDONE))
        return true
    }
}
