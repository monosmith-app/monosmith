package dev.monosmith.jetbrains.smithrun

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.execution.process.CapturingProcessHandler
import com.intellij.execution.process.ProcessAdapter
import com.intellij.execution.process.ProcessEvent
import com.intellij.execution.ui.ConsoleViewContentType
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Key
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
import java.nio.charset.StandardCharsets
import java.nio.file.Path

/** Runs the smith of the workspace and waits for it. Blocks: call it off the UI thread. */
object SmithRunClient {
    /**
     * Runs `smith <args>` in [root].
     *
     * @param ide the IDE project whose Output tab shows the run
     * @param root the workspace folder
     * @param args the arguments after `smith`
     * @param show print the command line and what smith prints in the Output tab, as it arrives
     * @return what smith printed and its exit code
     */
    fun run(
        ide: Project?,
        root: Path,
        args: List<String>,
        show: Boolean = false,
    ): SmithRunContract {
        val command = listOf(RunSmithCommandUseCase.smithFor(root)) + args
        val handler =
            CapturingProcessHandler(
                GeneralCommandLine(command).withWorkingDirectory(root).withCharset(StandardCharsets.UTF_8),
            )
        if (show && ide != null) {
            val console = OutputConsoleStore.consoleFor(ide)
            console.print("$ ${TaskCommandAlgorithm.display(command)}\n", ConsoleViewContentType.SYSTEM_OUTPUT)
            handler.addProcessListener(
                object : ProcessAdapter() {
                    override fun onTextAvailable(
                        event: ProcessEvent,
                        outputType: Key<*>,
                    ) = console.print(event.text, ConsoleViewContentType.getConsoleViewType(outputType))
                },
            )
            OutputConsoleStore.reveal(ide)
        }
        val output = handler.runProcess()
        return SmithRunContract(output.exitCode, output.stdout, output.stderr)
    }
}
