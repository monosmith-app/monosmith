package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.trycommand.TryCommandUseCase

/** Tries a smith command in a copy of the workspace, then offers to run it for real. */
class TryCommandActionHandler : SmithActionHandler({ ide, root -> TryCommandUseCase.start(ide, root) })
