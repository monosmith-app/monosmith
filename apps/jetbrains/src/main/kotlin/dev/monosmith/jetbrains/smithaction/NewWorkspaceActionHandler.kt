package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.ActionUpdateThread
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.DumbAware
import dev.monosmith.jetbrains.newworkspace.NewWorkspaceUseCase
import java.nio.file.Path

/** Creates a MonoSmith workspace in the folder the IDE has open: the defaults or the questions, then the files, then the write. */
class NewWorkspaceActionHandler :
    AnAction(),
    DumbAware {
    override fun getActionUpdateThread() = ActionUpdateThread.BGT

    override fun update(event: AnActionEvent) {
        event.presentation.isEnabledAndVisible = event.project?.basePath != null
    }

    override fun actionPerformed(event: AnActionEvent) {
        val ide = event.project ?: return
        val folder = ide.basePath?.let(Path::of) ?: return
        NewWorkspaceUseCase.start(ide, folder)
    }
}
