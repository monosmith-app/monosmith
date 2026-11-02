package dev.monosmith.jetbrains.smithaction

import dev.monosmith.jetbrains.readcommands.ReadSmithCommandUseCase

/** Shows the installed smith, the newest one and the settings of the workspace. */
class InfoActionHandler :
    SmithActionHandler({ ide, root -> ReadSmithCommandUseCase.start(ide, root, "the smith info", listOf("info")) })
