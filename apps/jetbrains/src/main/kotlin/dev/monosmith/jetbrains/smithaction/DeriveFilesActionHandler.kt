package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.derivefiles.DeriveFilesUseCase

/** Acts on the derived files that need action: explain, overwrite, keep my edit, eject, reclaim. */
class DeriveFilesActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.start(ide, root) })
