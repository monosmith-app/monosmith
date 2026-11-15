package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.cachecontrol.CacheControlUseCase

/** Deletes the local cache, after the consequence. */
class CacheClearActionHandler : SmithActionHandler({ ide, root -> CacheControlUseCase.start(ide, root, CacheControlUseCase.Kind.CLEAR) })
