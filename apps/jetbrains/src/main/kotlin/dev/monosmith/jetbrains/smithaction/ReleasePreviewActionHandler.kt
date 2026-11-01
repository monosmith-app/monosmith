package dev.monosmith.jetbrains.smithaction

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.smithcommand.ActionCommandAlgorithm
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
import java.nio.file.Path

/** `smith release --dry-run`: shows what a release would publish. It changes nothing. */
class ReleasePreviewActionHandler : SmithActionHandler() {
    override fun perform(
        event: AnActionEvent,
        ide: Project,
        root: Path,
    ) = RunSmithCommandUseCase.run(ide, root, "smith release --dry-run") { ActionCommandAlgorithm.releasePreview(it) }
}
