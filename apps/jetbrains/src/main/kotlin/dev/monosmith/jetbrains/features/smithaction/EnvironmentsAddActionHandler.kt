package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.workspacetools.WorkspaceToolsUseCase

/** Adds environments to a web app, after the plan. */
class EnvironmentsAddActionHandler :
    SmithActionHandler({
        ide,
        root,
        ->
        WorkspaceToolsUseCase.changeEnvironments(ide, root, remove = false, project = null)
    })
