package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.trycommand.TryCommandUseCase

/** Tries a smith command in a copy of the workspace, then offers to run it for real. */
class TryCommandActionHandler : SmithActionHandler({ ide, root -> TryCommandUseCase.start(ide, root) })
