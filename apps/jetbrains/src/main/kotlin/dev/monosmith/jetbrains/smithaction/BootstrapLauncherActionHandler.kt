package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Writes the smith launcher (`smith bootstrap`) after its plan. */
class BootstrapLauncherActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.writeLauncher(ide, root) })
