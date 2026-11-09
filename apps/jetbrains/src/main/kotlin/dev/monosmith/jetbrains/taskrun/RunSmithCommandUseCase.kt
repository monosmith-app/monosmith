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

/** Runs any smith command in a Run window, printing the command line first. */
object RunSmithCommandUseCase {
    /** Which smith the workspace uses: the one installed in it, else the one on the PATH. */
    fun smithFor(root: Path): String {
        val windows = System.getProperty("os.name").startsWith("Windows")
        return ResolveSmithExecutablePolicy.resolve(root, windows, Files::exists)
    }

    /**
     * Runs the command [build] makes in the workspace [root].
     *
     * @param title the name of the Run window
     * @param build makes the argument vector from the smith to run
     */
    fun run(
        ide: Project,
        root: Path,
        title: String,
        build: (smith: String) -> List<String>,
    ) {
        val command = build(smithFor(root))
        val console = TextConsoleBuilderFactory.getInstance().createBuilder(ide).console
        console.print("$ ${TaskCommandAlgorithm.display(command)}\n", ConsoleViewContentType.SYSTEM_OUTPUT)
        val handler =
            KillableProcessHandler(
                GeneralCommandLine(command).withWorkingDirectory(root).withCharset(StandardCharsets.UTF_8),
            )
        RunContentExecutor(ide, handler)
            .withConsole(console)
            .withTitle(title)
            .run()
    }
}
