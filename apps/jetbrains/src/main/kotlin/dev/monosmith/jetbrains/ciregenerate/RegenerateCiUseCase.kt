package dev.monosmith.jetbrains.ciregenerate

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import java.nio.file.Path

/**
 * Regenerates the CI pipelines (`smith ci generate`). The plan is `--dry-run`: a diff of every file it
 * rewrites or deletes. A pipeline the team wrote with `smith pipeline` is not touched. smith has no
 * `--commit` here, so nothing is committed.
 */
object RegenerateCiUseCase {
    private const val TITLE = "MonoSmith: regenerate the CI pipelines"

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
                    "The plan is a diff of every file smith rewrites or deletes. " +
                        "A pipeline of your own is not touched. Nothing was written yet.",
                okText = "Regenerate",
                offerCommit = false,
                apply = { listOf("ci", "generate", "--yes") },
                done = "The CI pipelines are regenerated",
            ),
        ) {}
}
