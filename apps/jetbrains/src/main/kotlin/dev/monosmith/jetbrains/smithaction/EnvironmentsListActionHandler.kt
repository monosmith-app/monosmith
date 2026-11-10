package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Lists the environments of the web apps. */
class EnvironmentsListActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.listEnvironments(ide, root) })
