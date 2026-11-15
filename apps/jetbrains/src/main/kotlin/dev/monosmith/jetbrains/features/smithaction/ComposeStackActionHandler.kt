package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.composestack.ComposeStackUseCase

/** Compose a stack: search everything smith can add, see conflicts, end in the plan. */
class ComposeStackActionHandler : SmithActionHandler({ ide, root -> ComposeStackUseCase.start(ide, root) })
