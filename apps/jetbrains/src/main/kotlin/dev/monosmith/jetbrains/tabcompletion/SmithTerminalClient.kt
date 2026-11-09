package dev.monosmith.jetbrains.tabcompletion

import com.intellij.openapi.project.Project
import org.jetbrains.plugins.terminal.TerminalToolWindowManager

/** Opens a MonoSmith tab in the terminal of the IDE and types a smith command there. */
object SmithTerminalClient {
    private const val TAB = "MonoSmith"

    /** The command that sets up tab completion; the person answers its question in the terminal. */
    const val INSTALL_COMMAND = "smith completion install"

    /**
     * Types [command] in a new terminal tab that starts in [folder]. It runs on the UI thread.
     *
     * @param ide the IDE project
     * @param folder where the terminal starts
     * @param command the line to run, as the person could type it
     */
    fun run(
        ide: Project,
        folder: String?,
        command: String,
    ) {
        TerminalToolWindowManager
            .getInstance(ide)
            .createShellWidget(folder, TAB, true, true)
            .sendCommandToExecute(command)
    }
}
