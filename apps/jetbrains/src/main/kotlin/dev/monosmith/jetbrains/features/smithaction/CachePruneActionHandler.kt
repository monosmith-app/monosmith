package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.cachecontrol.CacheControlUseCase

/** Trims the local cache to its limit, after the plan. */
class CachePruneActionHandler : SmithActionHandler({ ide, root -> CacheControlUseCase.start(ide, root, CacheControlUseCase.Kind.PRUNE) })
