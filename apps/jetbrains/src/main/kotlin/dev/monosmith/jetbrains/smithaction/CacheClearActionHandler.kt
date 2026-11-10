package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.cachecontrol.CacheControlUseCase

/** Deletes the local cache, after the consequence. */
class CacheClearActionHandler : SmithActionHandler({ ide, root -> CacheControlUseCase.start(ide, root, CacheControlUseCase.Kind.CLEAR) })
