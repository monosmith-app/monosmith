package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.workspacetools.WorkspaceToolsUseCase

/** Shows the divergences from the defaults (override, ejected, drift). */
class DivergencesActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.divergences(ide, root) })
