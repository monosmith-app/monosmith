package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Removes environments from a web app, after the plan; the .env files stay. */
class EnvironmentsRemoveActionHandler :
    SmithActionHandler({
        ide,
        root,
        ->
        WorkspaceToolsUseCase.changeEnvironments(ide, root, remove = true, project = null)
    })
