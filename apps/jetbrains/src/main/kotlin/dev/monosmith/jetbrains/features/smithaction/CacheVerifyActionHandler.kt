package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.readcommands.ReadSmithCommandUseCase
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_SMITHACTION_THE_CACHE_CHECK
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.say

/** Checks the local cache for damage. */
class CacheVerifyActionHandler :
    SmithActionHandler(
        { ide, root -> ReadSmithCommandUseCase.start(ide, root, say(JETBRAINS_SMITHACTION_THE_CACHE_CHECK), listOf("cache", "verify")) },
    )
