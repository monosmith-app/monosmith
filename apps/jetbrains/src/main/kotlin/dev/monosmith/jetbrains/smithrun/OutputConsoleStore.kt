package dev.monosmith.jetbrains.smithrun

import com.intellij.execution.filters.TextConsoleBuilderFactory
import com.intellij.execution.ui.ConsoleView
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Disposer
import com.intellij.openapi.wm.ToolWindowManager
import java.util.WeakHashMap

/** The output console of the MonoSmith tool window: one for each IDE project, kept for the session. */
object OutputConsoleStore {
    /** The id of the tool window. */
    const val TOOL_WINDOW_ID = "MonoSmith"

    /** The title of the tab that holds the console. */
    const val TAB_TITLE = "Output"

    private val consoles = WeakHashMap<Project, ConsoleView>()

    /** The console of [ide]; it is created on first use and lives as long as the IDE project. */
    @Synchronized
    fun consoleFor(ide: Project): ConsoleView =
        consoles.getOrPut(ide) {
            TextConsoleBuilderFactory
                .getInstance()
                .createBuilder(ide)
                .console
                .also { Disposer.register(ide, it) }
        }

    /** Shows the tool window on the tab [tab] (the Output tab by default), so the person sees what smith prints. */
    fun reveal(
        ide: Project,
        tab: String = TAB_TITLE,
    ) {
        ApplicationManager.getApplication().invokeLater {
            val window = ToolWindowManager.getInstance(ide).getToolWindow(TOOL_WINDOW_ID) ?: return@invokeLater
            window.show()
            window.contentManager.findContent(tab)?.let(window.contentManager::setSelectedContent)
        }
    }
}
