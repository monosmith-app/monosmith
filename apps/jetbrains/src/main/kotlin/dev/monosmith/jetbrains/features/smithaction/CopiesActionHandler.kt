package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.readcommands.ReadSmithCommandUseCase
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_SMITHACTION_THE_COPIES_BETWEEN_PROJECTS
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.say

/** Counts the files the projects keep as copies. */
class CopiesActionHandler :
    SmithActionHandler(
        { ide, root -> ReadSmithCommandUseCase.start(ide, root, say(JETBRAINS_SMITHACTION_THE_COPIES_BETWEEN_PROJECTS), listOf("copies")) },
    )
