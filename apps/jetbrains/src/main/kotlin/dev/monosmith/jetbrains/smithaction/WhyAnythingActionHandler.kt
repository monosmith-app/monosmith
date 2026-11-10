package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase

/** `smith why`: asks for a file, a project, a task or a setting path, then explains it. */
class WhyAnythingActionHandler : SmithActionHandler({ ide, root -> WorkspaceToolsUseCase.why(ide, root, null) })
