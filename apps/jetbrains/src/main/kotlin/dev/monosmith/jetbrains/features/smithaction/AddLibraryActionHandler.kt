package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.installdependency.InstallDependencyUseCase

/** Use a library of this workspace in a project (`smith install -p <project> workspace:<library>`). */
class AddLibraryActionHandler : SmithActionHandler({ ide, root -> InstallDependencyUseCase.start(ide, root, only = "workspace") })
