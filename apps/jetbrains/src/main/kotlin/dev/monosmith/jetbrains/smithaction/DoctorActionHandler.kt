package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.smithcommand.ActionCommandAlgorithm
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
import java.nio.file.Path

/** `smith doctor`: checks the workspace and says what to fix. It changes nothing. */
class DoctorActionHandler : SmithActionHandler() {
    override fun perform(
        event: AnActionEvent,
        ide: Project,
        root: Path,
    ) = RunSmithCommandUseCase.run(ide, root, "smith doctor") { ActionCommandAlgorithm.doctor(it, fix = false) }
}
