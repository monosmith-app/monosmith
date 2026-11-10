package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.derivefiles.DeriveFilesUseCase

/** Previews what a plain `smith derive` would write (`--dry-run`). */
class DerivePreviewActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.preview(ide, root) })
