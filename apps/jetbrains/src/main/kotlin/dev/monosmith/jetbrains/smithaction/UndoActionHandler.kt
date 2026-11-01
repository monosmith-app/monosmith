package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
import dev.monosmith.jetbrains.smithcommand.ActionCommandAlgorithm
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
import java.nio.file.Path

/** `smith undo`: reverts the last operation smith committed, after a question. */
class UndoActionHandler : SmithActionHandler() {
    override fun perform(
        event: AnActionEvent,
        ide: Project,
        root: Path,
    ) {
        val ask =
            Messages.showYesNoDialog(
                ide,
                "smith reverts the last change it committed with git revert: a new commit, and your history stays. " +
                    "smith refuses when a later commit or an uncommitted change touches the same files, and says why.",
                "MonoSmith: undo the last change?",
                "Undo",
                "Cancel",
                Messages.getQuestionIcon(),
            )
        if (ask != Messages.YES) return
        RunSmithCommandUseCase.run(ide, root, "smith undo") { ActionCommandAlgorithm.undo(it) }
    }
}
