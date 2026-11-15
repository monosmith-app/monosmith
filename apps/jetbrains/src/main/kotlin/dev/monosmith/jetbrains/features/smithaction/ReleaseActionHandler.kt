package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.releaseflow.ReleaseWorkspaceUseCase

/** Release: the plan, hold packages back, change a bump, then the last question (a release cannot be undone). */
class ReleaseActionHandler : SmithActionHandler({ ide, root -> ReleaseWorkspaceUseCase.start(ide, root) })
