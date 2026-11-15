package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.ciregenerate.RegenerateCiUseCase

/** Regenerates the CI pipelines, after the diff of every file. */
class RegenerateCiActionHandler : SmithActionHandler({ ide, root -> RegenerateCiUseCase.start(ide, root) })
