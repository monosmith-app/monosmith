package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.installtools.InstallToolsUseCase

/** Install the tools the workspace needs: one choice per tool with the command visible. */
class InstallToolsActionHandler : SmithActionHandler({ ide, root -> InstallToolsUseCase.start(ide, root) })
