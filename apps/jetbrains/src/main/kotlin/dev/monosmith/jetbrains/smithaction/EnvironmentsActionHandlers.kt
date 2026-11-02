package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Lists the environments of the web apps. */
class EnvironmentsListActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.listEnvironments(ide, root) })

/** Adds environments to a web app, after the plan. */
class EnvironmentsAddActionHandler :
    SmithActionHandler({
        ide,
        root,
        ->
        WorkspaceToolsUseCase.changeEnvironments(ide, root, remove = false, project = null)
    })

/** Removes environments from a web app, after the plan; the .env files stay. */
class EnvironmentsRemoveActionHandler :
    SmithActionHandler({
        ide,
        root,
        ->
        WorkspaceToolsUseCase.changeEnvironments(ide, root, remove = true, project = null)
    })
