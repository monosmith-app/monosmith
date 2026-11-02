package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase

/** Shows the status of the local cache. */
class CacheStatusActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the cache status", listOf("cache", "status")) })
