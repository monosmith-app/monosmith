package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHACTION_THE_TOOLS
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/** Lists the tools the workspace needs and the state of each. */
class ToolsListActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, say(JETBRAINS_SMITHACTION_THE_TOOLS), listOf("tools")) })
