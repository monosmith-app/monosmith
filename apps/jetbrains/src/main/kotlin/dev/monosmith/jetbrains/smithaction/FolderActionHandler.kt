package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.ActionUpdateThread
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.Project
import java.nio.file.Path

/**
 * The base of the actions that work on the folder the IDE has open, with or without a smith.toml: bringing
 * a repository in, converting a workspace.
 *
 * @param start starts the flow for the open folder; it runs on the UI thread and must not block
 */
open class FolderActionHandler(
    private val start: (ide: Project, folder: Path) -> Unit,
) : AnAction(),
    DumbAware {
    override fun getActionUpdateThread() = ActionUpdateThread.BGT

    override fun update(event: AnActionEvent) {
        event.presentation.isEnabledAndVisible = event.project?.basePath != null
    }

    override fun actionPerformed(event: AnActionEvent) {
        val ide = event.project ?: return
        val folder = ide.basePath?.let(Path::of) ?: return
        start(ide, folder)
    }
}
