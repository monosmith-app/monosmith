package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.releaseflow.ReleaseWorkspaceUseCase

/** `smith release --dry-run`: shows what a release would publish and changes nothing. */
class ReleasePreviewActionHandler : SmithActionHandler({ ide, root -> ReleaseWorkspaceUseCase.preview(ide, root) })
