package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.workspacetools.WorkspaceToolsUseCase

/** Lists the environments of the web apps. */
class EnvironmentsListActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.listEnvironments(ide, root) })
