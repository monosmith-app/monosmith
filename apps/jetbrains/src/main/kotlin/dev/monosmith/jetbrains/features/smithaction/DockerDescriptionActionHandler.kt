package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.dockerhub.DockerDescriptionUseCase

/** Sets the Docker Hub descriptions, after naming the repositories. */
class DockerDescriptionActionHandler : SmithActionHandler({ ide, root -> DockerDescriptionUseCase.start(ide, root) })
