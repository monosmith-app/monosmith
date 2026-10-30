package dev.monosmith.jetbrains.workspace

import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.execution.util.ExecUtil
import dev.monosmith.jetbrains.smithcommand.ResolveSmithExecutablePolicy
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import java.nio.charset.StandardCharsets
import java.nio.file.Files
import java.nio.file.Path

/** Asks the workspace's smith for its projects. Blocks: call it off the UI thread. */
object ListProjectsUseCase {
    fun list(root: Path): Result<List<SmithProjectContract>> =
        runCatching {
            val windows = System.getProperty("os.name").startsWith("Windows")
            val smith = ResolveSmithExecutablePolicy.resolve(root, windows, Files::exists)
            val output =
                ExecUtil.execAndGetOutput(
                    GeneralCommandLine(TaskCommandAlgorithm.listProjects(smith))
                        .withWorkingDirectory(root)
                        .withCharset(StandardCharsets.UTF_8),
                )
            check(output.exitCode == 0) { output.stderr.ifBlank { "smith projects exited with ${output.exitCode}" } }
            ProjectsJsonMapper.toProjects(output.stdout)
        }
}
