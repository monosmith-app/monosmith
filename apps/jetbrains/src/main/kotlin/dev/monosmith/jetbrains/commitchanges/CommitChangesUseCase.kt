package dev.monosmith.jetbrains.commitchanges

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Commits the changes of the work tree as one operation that Undo can revert (`smith commit`). It runs
 * `--dry-run`, shows the files, asks which ones go in (only the operation's files, everything git shows,
 * or cancel), and only then commits with `--yes`. smith never commits unasked: this runs only when the
 * person chose the action or pressed a button.
 */
object CommitChangesUseCase {
    /**
     * Starts the flow in a background task.
     *
     * @param operation the words that name the operation, such as "add app web"; empty for the last operation
     * @param then runs after a commit, for the follow-up of the caller
     */
    fun start(
        ide: Project,
        root: Path,
        operation: String = "",
        then: () -> Unit = {},
    ) = BackgroundFlowClient.start(ide, "MonoSmith: commit the changes") {
        if (run(IdePortsClient.of(ide, root), operation)) then()
    }

    /**
     * Runs the flow with the given ports.
     *
     * @param operation the words that name the operation, such as "add app web"; empty for the last operation
     * @return true when smith made the commit
     */
    fun run(
        ports: FlowPortsContract,
        operation: String = "",
    ): Boolean {
        val plan = ports.smith.run(CommitPlanAlgorithm.arguments(operation, null))
        if (!plan.ok) {
            ports.notice.warn("Nothing was committed. ${FailureTextMapper.describe(plan)}")
            return false
        }
        val read = CommitPlanAlgorithm.read(plan.stdout)
        val more = read.extras.size
        val extras = if (more == 0) "" else " Git also shows $more other changed files; they stay out unless you pick all."
        val header = "smith makes one commit that Undo can revert. Pick which files go in. Nothing is committed until you do.$extras"
        val answer =
            ports.prompt.plan(
                "MonoSmith: commit the changes",
                header,
                read.text,
                "Commit only the files of the operation",
                offerCommit = false,
                others = listOf("Commit everything git shows"),
            ) ?: return false
        val scope = if (answer.other == 0) "all" else "operation"
        val done = ports.smith.run(CommitPlanAlgorithm.arguments(operation, scope), show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn("Nothing was committed. ${FailureTextMapper.describe(done)}")
            return false
        }
        ports.notice.info(done.lines.lastOrNull() ?: "The changes are committed as one change. Undo reverts it.")
        return true
    }
}
