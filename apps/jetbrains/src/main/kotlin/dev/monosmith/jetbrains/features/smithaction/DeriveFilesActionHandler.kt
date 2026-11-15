package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.derivefiles.DeriveFilesUseCase

/** Acts on the derived files that need action: explain, overwrite, keep my edit, eject, reclaim. */
class DeriveFilesActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.start(ide, root) })
