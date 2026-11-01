package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.ActionUpdateThread
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
import dev.monosmith.jetbrains.workspace.FindWorkspaceRootUseCase
import java.nio.file.Files
import java.nio.file.Path

/**
 * The base of the MonoSmith menu actions. It finds the workspace of the IDE project (the nearest
 * folder up with a smith.toml) and says so when there is none, so each action only says what to run.
 */
abstract class SmithActionHandler :
    AnAction(),
    DumbAware {
    override fun getActionUpdateThread() = ActionUpdateThread.BGT

    override fun update(event: AnActionEvent) {
        event.presentation.isEnabledAndVisible = event.project != null
    }

    final override fun actionPerformed(event: AnActionEvent) {
        val ide = event.project ?: return
        val root = ide.basePath?.let(Path::of)?.let { FindWorkspaceRootUseCase.find(it, Files::exists) }
        if (root == null) {
            Messages.showInfoMessage(
                ide,
                "There is no smith.toml in this project or above it, so there is no MonoSmith workspace to work on.",
                "MonoSmith",
            )
            return
        }
        perform(event, ide, root)
    }

    /** Does the action in the workspace at [root]. Runs on the UI thread. */
    protected abstract fun perform(
        event: AnActionEvent,
        ide: Project,
        root: Path,
    )
}
