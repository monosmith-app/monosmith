package dev.monosmith.jetbrains.dockerhub

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_PARITY_FORMS_SET_THE_DESCRIPTIONS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DOCKERHUB_MONOSMITH_DOCKER_HUB_DESCRIPTIONS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DOCKERHUB_SMITH_CALLS_DOCKER_HUB_WITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DOCKERHUB_THE_DOCKER_HUB_DESCRIPTIONS_ARE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Sets the Docker Hub descriptions of the apps whose registries include docker.io (`smith docker
 * description`). The plan names the repositories without calling Docker Hub (`--dry-run`); only a yes calls
 * it, with the person's own login. Nothing in the workspace changes, so nothing is committed.
 */
object DockerDescriptionUseCase {
    private val TITLE = say(JETBRAINS_DOCKERHUB_MONOSMITH_DOCKER_HUB_DESCRIPTIONS)

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
                header = say(JETBRAINS_DOCKERHUB_SMITH_CALLS_DOCKER_HUB_WITH),
                okText = say(GUIDE_PARITY_FORMS_SET_THE_DESCRIPTIONS),
                offerCommit = false,
                apply = { listOf("docker", "description") },
                done = say(JETBRAINS_DOCKERHUB_THE_DOCKER_HUB_DESCRIPTIONS_ARE),
            ),
        ) {}
}
