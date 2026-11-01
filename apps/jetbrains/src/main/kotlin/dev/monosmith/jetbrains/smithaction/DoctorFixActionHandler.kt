package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
import dev.monosmith.jetbrains.smithcommand.ActionCommandAlgorithm
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
import java.nio.file.Path

/**
 * `smith doctor --fix`: shows the fixes doctor can make as diffs and applies the safe ones. A fix
 * with a question needs a terminal, so the output names it and the Run window leaves it alone.
 */
class DoctorFixActionHandler : SmithActionHandler() {
    override fun perform(
        event: AnActionEvent,
        ide: Project,
        root: Path,
    ) {
        val ask =
            Messages.showYesNoDialog(
                ide,
                "smith doctor --fix shows each fix as a diff and applies the safe ones. Review the changes with git afterwards.",
                "MonoSmith: apply the safe fixes?",
                "Fix",
                "Cancel",
                Messages.getQuestionIcon(),
            )
        if (ask != Messages.YES) return
        RunSmithCommandUseCase.run(ide, root, "smith doctor --fix") { ActionCommandAlgorithm.doctor(it, fix = true) }
    }
}
