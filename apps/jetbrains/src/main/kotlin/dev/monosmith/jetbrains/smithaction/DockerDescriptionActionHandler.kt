package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.dockerhub.DockerDescriptionUseCase

/** Sets the Docker Hub descriptions, after naming the repositories. */
class DockerDescriptionActionHandler : SmithActionHandler({ ide, root -> DockerDescriptionUseCase.start(ide, root) })
