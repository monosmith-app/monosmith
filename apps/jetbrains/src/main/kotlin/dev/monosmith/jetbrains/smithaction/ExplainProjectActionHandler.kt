package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Explains how the tasks of a project resolved (`smith explain`). */
class ExplainProjectActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.explain(ide, root, null) })
