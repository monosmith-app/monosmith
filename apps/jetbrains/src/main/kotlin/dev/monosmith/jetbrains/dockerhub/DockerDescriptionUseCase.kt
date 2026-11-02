package dev.monosmith.jetbrains.dockerhub

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import java.nio.file.Path

/**
 * Sets the Docker Hub descriptions of the apps whose registries include docker.io (`smith docker
 * description`). The plan names the repositories without calling Docker Hub (`--dry-run`); only a yes calls
 * it, with the person's own login. Nothing in the workspace changes, so nothing is committed.
 */
object DockerDescriptionUseCase {
    private const val TITLE = "MonoSmith: Docker Hub descriptions"

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) }

    /**
     * @return true when smith set the descriptions
     */
    fun run(ports: FlowPortsContract): Boolean =
        PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = listOf("docker", "description", "--dry-run"),
                header = "smith calls Docker Hub with your login and replaces the description of each image. Nothing was called yet.",
                okText = "Set the descriptions",
                offerCommit = false,
                apply = { listOf("docker", "description") },
                done = "The Docker Hub descriptions are set",
            ),
        ) {}
}
