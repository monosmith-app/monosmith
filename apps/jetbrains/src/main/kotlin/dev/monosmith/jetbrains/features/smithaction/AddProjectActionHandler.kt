package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.addproject.AddProjectUseCase
import dev.monosmith.jetbrains.features.composestack.ComposeStackUseCase

/** Add a project: outcome first, choices per axis, the plan from `--dry-run`, the unticked commit box. */
class AddProjectActionHandler :
    SmithActionHandler({ ide, root -> AddProjectUseCase.start(ide, root) { ComposeStackUseCase.start(ide, root) } })
