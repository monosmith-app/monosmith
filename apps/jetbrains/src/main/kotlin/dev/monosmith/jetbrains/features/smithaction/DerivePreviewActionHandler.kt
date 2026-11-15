package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.derivefiles.DeriveFilesUseCase

/** Previews what a plain `smith derive` would write (`--dry-run`). */
class DerivePreviewActionHandler : SmithActionHandler({ ide, root -> DeriveFilesUseCase.preview(ide, root) })
