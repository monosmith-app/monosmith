package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.cachecontrol.CacheControlUseCase

/** Trims the local cache to its limit, after the plan. */
class CachePruneActionHandler : SmithActionHandler({ ide, root -> CacheControlUseCase.start(ide, root, CacheControlUseCase.Kind.PRUNE) })
