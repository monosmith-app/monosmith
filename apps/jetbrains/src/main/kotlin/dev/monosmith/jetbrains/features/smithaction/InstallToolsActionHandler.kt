package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.installtools.InstallToolsUseCase

/** Install the tools the workspace needs: one choice per tool with the command visible. */
class InstallToolsActionHandler : SmithActionHandler({ ide, root -> InstallToolsUseCase.start(ide, root) })
