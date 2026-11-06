package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHACTION_THE_CACHE_CHECK
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** Checks the local cache for damage. */
class CacheVerifyActionHandler :
    SmithActionHandler(
        { ide, root -> ReadSmithCommandUseCase.start(ide, root, say(JETBRAINS_SMITHACTION_THE_CACHE_CHECK), listOf("cache", "verify")) },
    )
