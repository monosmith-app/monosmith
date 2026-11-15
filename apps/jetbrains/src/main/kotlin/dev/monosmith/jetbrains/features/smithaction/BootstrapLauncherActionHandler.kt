package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.workspacetools.WorkspaceToolsUseCase

/** Writes the smith launcher (`smith bootstrap`) after its plan. */
class BootstrapLauncherActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.writeLauncher(ide, root) })
