package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase

/** Says why a task would run again instead of coming from the cache. */
class WhyMissActionHandler : SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.startWhyMiss(ide, root) })
