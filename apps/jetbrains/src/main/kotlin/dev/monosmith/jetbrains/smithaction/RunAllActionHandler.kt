package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.runacross.RunAcrossUseCase

/** Runs one task in every project that has it. */
class RunAllActionHandler : SmithActionHandler({ ide, root -> RunAcrossUseCase.start(ide, root, affected = false) })
