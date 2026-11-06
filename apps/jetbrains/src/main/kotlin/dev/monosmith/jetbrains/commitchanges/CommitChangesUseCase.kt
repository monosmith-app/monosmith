package dev.monosmith.jetbrains.commitchanges

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMMITCHANGES_COMMIT_EVERYTHING_GIT_SHOWS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMMITCHANGES_COMMIT_ONLY_THE_FILES_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMMITCHANGES_GIT_ALSO_SHOWS_OTHER_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMMITCHANGES_MONOSMITH_COMMIT_THE_CHANGES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMMITCHANGES_NOTHING_WAS_COMMITTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMMITCHANGES_SMITH_MAKES_ONE_COMMIT_THAT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMMITCHANGES_THE_CHANGES_ARE_COMMITTED_AS
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_COMMITCHANGES_MONOSMITH_COMMIT_THE_CHANGES)) {
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
            ports.notice.warn(say(JETBRAINS_COMMITCHANGES_NOTHING_WAS_COMMITTED, FailureTextMapper.describe(plan).toString()))
            return false
        }
        val read = CommitPlanAlgorithm.read(plan.stdout)
        val more = read.extras.size
        val extras = if (more == 0) "" else say(JETBRAINS_COMMITCHANGES_GIT_ALSO_SHOWS_OTHER_CHANGED, more.toString())
        val header = say(JETBRAINS_COMMITCHANGES_SMITH_MAKES_ONE_COMMIT_THAT, extras.toString())
        val answer =
            ports.prompt.plan(
                say(JETBRAINS_COMMITCHANGES_MONOSMITH_COMMIT_THE_CHANGES),
                header,
                read.text,
                say(JETBRAINS_COMMITCHANGES_COMMIT_ONLY_THE_FILES_OF),
                offerCommit = false,
                others = listOf(say(JETBRAINS_COMMITCHANGES_COMMIT_EVERYTHING_GIT_SHOWS)),
            ) ?: return false
        val scope = if (answer.other == 0) "all" else "operation"
        val done = ports.smith.run(CommitPlanAlgorithm.arguments(operation, scope), show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(say(JETBRAINS_COMMITCHANGES_NOTHING_WAS_COMMITTED, FailureTextMapper.describe(done).toString()))
            return false
        }
        ports.notice.info(done.lines.lastOrNull() ?: say(JETBRAINS_COMMITCHANGES_THE_CHANGES_ARE_COMMITTED_AS))
        return true
    }
}
