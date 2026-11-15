package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.syncversions.SyncVersionsUseCase

/** `smith sync`: which version wins, then the plan, then the write. */
class SyncVersionsActionHandler : SmithActionHandler({ ide, root -> SyncVersionsUseCase.start(ide, root) })
