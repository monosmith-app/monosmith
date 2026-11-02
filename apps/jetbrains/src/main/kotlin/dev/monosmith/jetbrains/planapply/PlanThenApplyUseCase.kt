package dev.monosmith.jetbrains.planapply

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
import java.nio.file.Path

/**
 * Shows what smith would do (the plan, or a fixed text when smith has no plan command), asks, and only on
 * a yes runs the change with `--yes`. smith commits only when the unticked commit box is ticked; after a
 * change that is not committed, a button offers the Commit action.
 */
object PlanThenApplyUseCase {
    /** Starts the step in a background task. */
    fun start(
        ide: Project,
        root: Path,
        step: PlanStepContract,
    ) = BackgroundFlowClient.start(ide, step.title) { run(ide, root, step) }

    /**
     * Runs the step in the current background flow.
     *
     * @return true when smith made the change
     */
    fun run(
        ide: Project,
        root: Path,
        step: PlanStepContract,
    ): Boolean {
        val plan =
            if (step.planArgs == null) {
                step.fixedPlan
            } else {
                val planned = SmithRunClient.run(ide, root, step.planArgs)
                if (!planned.ok) {
                    NoticeClient.warn(ide, "Nothing was changed. ${FailureTextMapper.describe(planned)}")
                    return false
                }
                planned.stdout.trim()
            }
        val answer =
            PromptClient.plan(ide, step.title, step.header, plan, step.okText, step.offerCommit, others = step.others) ?: return false
        val args = step.apply(answer) ?: return false
        val committed = answer.commit && step.offerCommit
        val done = SmithRunClient.run(ide, root, if (committed) args + "--commit" else args, show = true)
        WorkspaceChangedStore.changed(ide)
        if (!done.ok) {
            NoticeClient.warn(ide, "Nothing was changed. ${FailureTextMapper.describe(done)}")
            return false
        }
        val buttons =
            if (committed || !step.offerCommit) {
                emptyList()
            } else {
                listOf(
                    NoticeButtonContract("Commit changes…", "Make one commit that Undo can revert; you choose the files first.") {
                        CommitChangesUseCase.start(ide, root, step.operation)
                    },
                )
            }
        val tail =
            if (committed) {
                "Committed as one change: Undo reverts it."
            } else if (step.offerCommit) {
                "Not committed: review it with the Commit window, or commit it now."
            } else {
                "Nothing is committed."
            }
        NoticeClient.info(ide, "${step.done}. $tail", buttons)
        return true
    }
}
