package dev.monosmith.jetbrains.taskrun

import com.intellij.execution.RunContentExecutor
import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.execution.filters.TextConsoleBuilderFactory
import com.intellij.execution.process.KillableProcessHandler
import com.intellij.execution.ui.ConsoleViewContentType
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.smithcommand.ResolveSmithExecutablePolicy
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import java.nio.charset.StandardCharsets
import java.nio.file.Files
import java.nio.file.Path

/** Runs `smith run <task> -p <project>` in a Run window, printing the command line first. */
object RunTaskUseCase {
    fun run(
        ide: Project,
        root: Path,
        project: String,
        task: String,
    ) {
        val windows = System.getProperty("os.name").startsWith("Windows")
        val smith = ResolveSmithExecutablePolicy.resolve(root, windows, Files::exists)
        val command = TaskCommandAlgorithm.runTask(smith, project, task)
        val console = TextConsoleBuilderFactory.getInstance().createBuilder(ide).console
        console.print("$ ${TaskCommandAlgorithm.display(command)}\n", ConsoleViewContentType.SYSTEM_OUTPUT)
        val handler =
            KillableProcessHandler(
                GeneralCommandLine(command).withWorkingDirectory(root).withCharset(StandardCharsets.UTF_8),
            )
        RunContentExecutor(ide, handler)
            .withConsole(console)
            .withTitle("smith $task ($project)")
            .run()
    }
}
