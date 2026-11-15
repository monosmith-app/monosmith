package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.workspacetools.WorkspaceToolsUseCase

/** `smith why`: asks for a file, a project, a task or a setting path, then explains it. */
class WhyAnythingActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.why(ide, root, null) })
