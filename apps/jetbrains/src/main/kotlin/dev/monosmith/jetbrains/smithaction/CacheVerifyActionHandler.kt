package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase

/** Checks the local cache for damage. */
class CacheVerifyActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the cache check", listOf("cache", "verify")) })
