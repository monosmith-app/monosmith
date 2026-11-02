package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase

/** Counts the files the projects keep as copies. */
class CopiesActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the copies between projects", listOf("copies")) })
