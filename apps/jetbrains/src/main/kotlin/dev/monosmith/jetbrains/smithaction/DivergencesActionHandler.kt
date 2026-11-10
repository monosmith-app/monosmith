package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** Shows the divergences from the defaults (override, ejected, drift). */
class DivergencesActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.divergences(ide, root) })
