package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Adds environments to a web app, after the plan. */
class EnvironmentsAddActionHandler :
    SmithActionHandler({
        ide,
        root,
        ->
        WorkspaceToolsUseCase.changeEnvironments(ide, root, remove = false, project = null)
    })
