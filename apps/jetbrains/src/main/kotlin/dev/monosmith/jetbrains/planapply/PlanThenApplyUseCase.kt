package dev.monosmith.jetbrains.planapply

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
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
     * Runs the step in the current background flow, in the IDE.
     *
     * @return true when smith made the change
     */
    fun run(
        ide: Project,
        root: Path,
        step: PlanStepContract,
    ): Boolean = run(IdePortsClient.of(ide, root), step) { CommitChangesUseCase.start(ide, root, it) }

    /**
     * Runs the step with the given ports.
     *
     * @param commitLater starts the Commit action for an operation; the notice button calls it
     * @return true when smith made the change
     */
    fun run(
        ports: FlowPortsContract,
        step: PlanStepContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val plan =
            if (step.planArgs == null) {
                step.fixedPlan
            } else {
                val planned = ports.smith.run(step.planArgs)
                if (!planned.ok) {
                    ports.notice.warn("Nothing was changed. ${FailureTextMapper.describe(planned)}")
                    return false
                }
                planned.stdout.trim()
            }
        val answer = ports.prompt.plan(step.title, step.header, plan, step.okText, step.offerCommit, others = step.others) ?: return false
        val args = step.apply(answer) ?: return false
        val committed = answer.commit && step.offerCommit
        val done = ports.smith.run(if (committed) args + "--commit" else args, show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn("Nothing was changed. ${FailureTextMapper.describe(done)}")
            return false
        }
        val buttons =
            if (committed || !step.offerCommit) {
                emptyList()
            } else {
                listOf(
                    NoticeButtonContract("Commit changes…", "Make one commit that Undo can revert; you choose the files first.") {
                        commitLater(step.operation)
                    },
                )
            }
        val tail =
            when {
                committed -> "Committed as one change: Undo reverts it."
                step.offerCommit -> "Not committed: review it with the Commit window, or commit it now."
                else -> "Nothing is committed."
            }
        ports.notice.info("${step.done}. $tail", buttons)
        return true
    }
}
