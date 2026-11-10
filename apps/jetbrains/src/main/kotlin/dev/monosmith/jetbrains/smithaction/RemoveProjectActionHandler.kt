package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.projectactions.ProjectActionsUseCase

/** Remove a project: the plan, then keep a copy or delete. */
class RemoveProjectActionHandler : SmithActionHandler({ ide, root -> ProjectActionsUseCase.remove(ide, root, null) })
