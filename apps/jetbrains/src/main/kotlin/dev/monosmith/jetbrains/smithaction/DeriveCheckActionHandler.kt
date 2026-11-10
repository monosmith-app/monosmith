package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.derivefiles.DeriveFilesUseCase

/** Checks the derived files (`smith derive --check`). */
class DeriveCheckActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.check(ide, root) })
