package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.readcommands.ReadSmithCommandUseCase
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_SMITHACTION_THE_CACHE_STATUS
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.say

/** Shows the status of the local cache. */
class CacheStatusActionHandler :
    SmithActionHandler(
        { ide, root -> ReadSmithCommandUseCase.start(ide, root, say(JETBRAINS_SMITHACTION_THE_CACHE_STATUS), listOf("cache", "status")) },
    )
