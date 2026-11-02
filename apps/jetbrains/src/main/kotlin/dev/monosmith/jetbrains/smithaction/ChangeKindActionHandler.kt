package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.projectactions.ProjectActionsUseCase

/** Make a project an app, a library or a package: the plan of `smith kind --dry-run`, then the change. */
class ChangeKindActionHandler : SmithActionHandler({ ide, root -> ProjectActionsUseCase.changeKind(ide, root, null) })
