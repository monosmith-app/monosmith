package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.installdependency.InstallDependencyUseCase

/** Use a library of this workspace in a project (`smith install -w <project> workspace:<library>`). */
class AddLibraryActionHandler : SmithActionHandler({ ide, root -> InstallDependencyUseCase.start(ide, root, only = "workspace") })
