package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.cachecontrol.CacheControlUseCase

/** Removes what is broken in the local cache, after the consequence. */
class CacheRepairActionHandler : SmithActionHandler({ ide, root -> CacheControlUseCase.start(ide, root, CacheControlUseCase.Kind.REPAIR) })
