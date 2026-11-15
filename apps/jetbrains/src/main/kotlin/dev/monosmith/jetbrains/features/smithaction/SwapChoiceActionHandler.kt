package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.swapchoice.SwapChoiceUseCase

/** Swap one choice of a project (its test runner, its style): the plan first, the commit box unticked. */
class SwapChoiceActionHandler : SmithActionHandler({ ide, root -> SwapChoiceUseCase.start(ide, root, null) })
