package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.installdependency.InstallDependencyUseCase

/** Install packages from the registry in a project (`smith install -p <project> --registry <packages>`). */
class InstallPackageActionHandler : SmithActionHandler({ ide, root -> InstallDependencyUseCase.start(ide, root, only = "registry") })
