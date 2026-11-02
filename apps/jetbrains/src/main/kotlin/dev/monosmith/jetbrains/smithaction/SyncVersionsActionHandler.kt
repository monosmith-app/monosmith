package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.syncversions.SyncVersionsUseCase

/** `smith sync`: which version wins, then the plan, then the write. */
class SyncVersionsActionHandler : SmithActionHandler({ ide, root -> SyncVersionsUseCase.start(ide, root) })
