package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.releaseflow.ReleaseWorkspaceUseCase

/** Release: the plan, hold packages back, change a bump, then the last question (a release cannot be undone). */
class ReleaseActionHandler : SmithActionHandler({ ide, root -> ReleaseWorkspaceUseCase.start(ide, root) })
