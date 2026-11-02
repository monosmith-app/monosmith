package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.installdependency.InstallDependencyUseCase

/** Install packages from the registry in a project (`smith install -w <project> --registry <packages>`). */
class InstallPackageActionHandler : SmithActionHandler({ ide, root -> InstallDependencyUseCase.start(ide, root, only = "registry") })
