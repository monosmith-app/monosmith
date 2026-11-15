package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.derivefiles.DeriveFilesUseCase

/** Checks the derived files (`smith derive --check`). */
class DeriveCheckActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.check(ide, root) })
