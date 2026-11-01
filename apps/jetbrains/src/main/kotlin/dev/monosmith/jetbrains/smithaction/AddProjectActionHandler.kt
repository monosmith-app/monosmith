package dev.monosmith.jetbrains.smithaction

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.execution.util.ExecUtil
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
import dev.monosmith.jetbrains.smithcommand.ActionCommandAlgorithm
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
import dev.monosmith.jetbrains.workspace.KindsJsonMapper
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.ProblemTextMapper
import dev.monosmith.jetbrains.workspace.SmithKindContract
import java.nio.charset.StandardCharsets
import java.nio.file.Path

/**
 * `smith add`: asks the kind, the name and whether to commit, then runs `smith add` in a Run window.
 * smith installs and tests the new project, so it takes a while; the Run window shows it.
 */
class AddProjectActionHandler : SmithActionHandler() {
    override fun perform(
        event: AnActionEvent,
        ide: Project,
        root: Path,
    ) {
        ApplicationManager.getApplication().executeOnPooledThread {
            val kinds = readKinds(root)
            val existing = ListProjectsUseCase.list(root).getOrDefault(emptyList()).map { it.name }
            ApplicationManager.getApplication().invokeLater {
                kinds.fold(
                    onSuccess = { ask(ide, root, it, existing) },
                    onFailure = { Messages.showWarningDialog(ide, "smith could not list what it can add. ${it.message}", "MonoSmith") },
                )
            }
        }
    }

    private fun ask(
        ide: Project,
        root: Path,
        kinds: List<SmithKindContract>,
        existing: List<String>,
    ) {
        val dialog = AddProjectDialogHandler(ide, kinds, existing)
        if (!dialog.showAndGet()) return
        val kind = dialog.kind
        val name = dialog.projectName
        RunSmithCommandUseCase.run(ide, root, "smith add ${kind.role()} $name") {
            ActionCommandAlgorithm.add(it, kind.role(), name, kind.language, dialog.commit)
        }
    }

    /** Blocks: call it off the UI thread. */
    private fun readKinds(root: Path): Result<List<SmithKindContract>> =
        runCatching {
            val output =
                ExecUtil.execAndGetOutput(
                    GeneralCommandLine(ActionCommandAlgorithm.listKinds(RunSmithCommandUseCase.smithFor(root)))
                        .withWorkingDirectory(root)
                        .withCharset(StandardCharsets.UTF_8),
                )
            check(output.exitCode == 0) { ProblemTextMapper.toLine(output.stderr).ifBlank { "smith kinds exited with ${output.exitCode}" } }
            KindsJsonMapper.toKinds(output.stdout)
        }
}
