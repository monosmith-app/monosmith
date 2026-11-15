package dev.monosmith.jetbrains.features.smithaction

import dev.monosmith.jetbrains.features.tabcompletion.SmithTerminalClient

/** Sets up tab completion: types `smith completion install` in the terminal, where the person answers its question. */
class TabCompletionActionHandler :
    FolderActionHandler({ ide, folder -> SmithTerminalClient.run(ide, folder.toString(), SmithTerminalClient.INSTALL_COMMAND) })
