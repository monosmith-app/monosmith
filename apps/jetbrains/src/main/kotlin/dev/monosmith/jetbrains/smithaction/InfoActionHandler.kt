package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHACTION_THE_SMITH_INFO
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** Shows the installed smith, the newest one and the settings of the workspace. */
class InfoActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, say(JETBRAINS_SMITHACTION_THE_SMITH_INFO), listOf("info")) })
