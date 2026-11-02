package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.ciregenerate.RegenerateCiUseCase

/** Regenerates the CI pipelines, after the diff of every file. */
class RegenerateCiActionHandler : SmithActionHandler({ ide, root -> RegenerateCiUseCase.start(ide, root) })
