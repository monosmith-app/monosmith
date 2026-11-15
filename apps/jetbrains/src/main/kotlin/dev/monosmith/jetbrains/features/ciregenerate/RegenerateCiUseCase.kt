package dev.monosmith.jetbrains.features.ciregenerate

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.features.flowports.FlowPortsContract
import dev.monosmith.jetbrains.features.flowports.IdePortsClient
import dev.monosmith.jetbrains.features.planapply.PlanStepContract
import dev.monosmith.jetbrains.features.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.features.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_CIREGENERATE_MONOSMITH_REGENERATE_THE_CI_PIPELINES
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_CIREGENERATE_REGENERATE
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_CIREGENERATE_THE_CI_PIPELINES_ARE_REGENERATED
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_CIREGENERATE_THE_PLAN_IS_A_DIFF
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Regenerates the CI pipelines (`smith ci generate`). The plan is `--dry-run`: a diff of every file it
 * rewrites or deletes. A pipeline the team wrote with `smith pipeline` is not touched. smith has no
 * `--commit` here, so nothing is committed.
 */
object RegenerateCiUseCase {
    private val TITLE = say(JETBRAINS_CIREGENERATE_MONOSMITH_REGENERATE_THE_CI_PIPELINES)

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) }

    /**
     * @return true when smith regenerated the pipelines
     */
    fun run(ports: FlowPortsContract): Boolean =
        PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = listOf("ci", "generate", "--dry-run"),
                header =
                    say(JETBRAINS_CIREGENERATE_THE_PLAN_IS_A_DIFF),
                okText = say(JETBRAINS_CIREGENERATE_REGENERATE),
                offerCommit = false,
                apply = { listOf("ci", "generate", "--yes") },
                done = say(JETBRAINS_CIREGENERATE_THE_CI_PIPELINES_ARE_REGENERATED),
            ),
        ) {}
}
