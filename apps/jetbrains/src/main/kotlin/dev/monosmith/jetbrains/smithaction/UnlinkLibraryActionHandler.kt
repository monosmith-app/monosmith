package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.projectactions.ProjectActionsUseCase

/** Unlink a library of this workspace from a project; the library stays. */
class UnlinkLibraryActionHandler : SmithActionHandler({ ide, root -> ProjectActionsUseCase.unlink(ide, root, null) })
