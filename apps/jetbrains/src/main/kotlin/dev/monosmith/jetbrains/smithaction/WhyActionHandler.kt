package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHACTION_IS_OUTSIDE_THE_MONOSMITH_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHACTION_MONOSMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import dev.monosmith.jetbrains.workspacetools.WorkspaceToolsUseCase
import java.nio.file.Path

/** `smith why <file>`: explains the file selected in the editor or the project view. */
class WhyActionHandler : SmithActionHandler() {
    override fun update(event: AnActionEvent) {
        event.presentation.isEnabledAndVisible =
            event.project != null && event.getData(CommonDataKeys.VIRTUAL_FILE)?.isDirectory == false
    }

    override fun perform(
        event: AnActionEvent,
        ide: Project,
        root: Path,
    ) {
        val file = event.getData(CommonDataKeys.VIRTUAL_FILE) ?: return
        val relative = runCatching { root.relativize(Path.of(file.path)).toString().replace('\\', '/') }.getOrNull()
        if (relative == null || relative.startsWith("..")) {
            Messages.showInfoMessage(
                ide,
                say(JETBRAINS_SMITHACTION_IS_OUTSIDE_THE_MONOSMITH_WORKSPACE, file.name.toString(), root.toString()),
                say(JETBRAINS_SMITHACTION_MONOSMITH),
            )
            return
        }
        WorkspaceToolsUseCase.why(ide, root, relative)
    }
}
