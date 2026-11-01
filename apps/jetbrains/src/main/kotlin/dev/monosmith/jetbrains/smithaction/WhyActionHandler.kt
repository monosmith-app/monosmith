package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
import dev.monosmith.jetbrains.smithcommand.ActionCommandAlgorithm
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
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
            Messages.showInfoMessage(ide, "${file.name} is outside the MonoSmith workspace at $root, so smith has nothing to say about it.", "MonoSmith")
            return
        }
        RunSmithCommandUseCase.run(ide, root, "smith why $relative") { ActionCommandAlgorithm.why(it, relative) }
    }
}
