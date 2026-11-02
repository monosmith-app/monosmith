package dev.monosmith.jetbrains.agentfile

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import java.nio.file.Path

/**
 * Writes AGENTS.md from the rules of the workspace (`smith agent file`), for the coding agents that work in
 * it. The plan is `--dry-run`: the files it would write. Text the team wrote outside smith's markers stays.
 * smith has no `--commit` here, so nothing is committed.
 */
object WriteAgentFileUseCase {
    private const val TITLE = "MonoSmith: write the agent file"

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) }

    /**
     * @return true when smith wrote the files
     */
    fun run(ports: FlowPortsContract): Boolean =
        PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = listOf("agent", "file", "--dry-run"),
                header = "Your own text outside smith's markers is kept. Nothing was written yet.",
                okText = "Write the files",
                offerCommit = false,
                apply = { listOf("agent", "file", "--yes") },
                done = "AGENTS.md is written",
            ),
        ) {}
}
