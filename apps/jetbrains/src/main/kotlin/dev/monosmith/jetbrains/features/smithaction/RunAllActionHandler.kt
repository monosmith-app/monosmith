package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.runacross.RunAcrossUseCase

/** Runs one task in every project that has it. */
class RunAllActionHandler : SmithActionHandler({ ide, root -> RunAcrossUseCase.start(ide, root, affected = false) })
