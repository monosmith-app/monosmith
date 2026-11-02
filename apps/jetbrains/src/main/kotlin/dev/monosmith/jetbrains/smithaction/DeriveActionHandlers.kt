package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.derivefiles.DeriveFilesUseCase

/** Acts on the derived files that need action: explain, overwrite, keep my edit, eject, reclaim. */
class DeriveFilesActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.start(ide, root) })

/** Checks the derived files (`smith derive --check`). */
class DeriveCheckActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.check(ide, root) })

/** Previews what a plain `smith derive` would write (`--dry-run`). */
class DerivePreviewActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.preview(ide, root) })
