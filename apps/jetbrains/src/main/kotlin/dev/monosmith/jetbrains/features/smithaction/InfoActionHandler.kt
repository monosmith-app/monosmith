package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.readcommands.ReadSmithCommandUseCase
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_SMITHACTION_THE_SMITH_INFO
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.say

/** Shows the installed smith, the newest one and the settings of the workspace. */
class InfoActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, say(JETBRAINS_SMITHACTION_THE_SMITH_INFO), listOf("info")) })
