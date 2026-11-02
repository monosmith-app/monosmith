package dev.monosmith.jetbrains.smithversion

import com.intellij.openapi.ide.CopyPasteManager
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import java.awt.datatransfer.StringSelection
import java.nio.file.Path

/**
 * Checks that the installed smith answers, and says how to get or update it when it does not. The plugin
 * runs the smith installed in the workspace or on the PATH and never installs one itself (ADR 0025), so the
 * notice offers the command to copy.
 */
object SmithPresenceUseCase {
    /** The command that installs or updates smith. */
    const val INSTALL_COMMAND = "npm i -g monosmith@latest"

    /** Runs `smith --version` and, when smith does not start or does not know a command the plugin needs, tells the person. Blocks. */
    fun check(
        ide: Project,
        root: Path,
    ) {
        val problem =
            runCatching { SmithRunClient.run(null, root, listOf("--version")) }.fold(
                onSuccess = { if (it.ok) null else "smith answered with an error. ${it.stderr.lines().firstOrNull().orEmpty()}" },
                onFailure = { "smith is not installed here, or it is not on the PATH." },
            ) ?: return
        NoticeClient.warn(
            ide,
            "$problem MonoSmith needs smith to list the projects and run the flows. Install or update it with: $INSTALL_COMMAND",
            listOf(
                NoticeButtonContract("Copy command", "Copy $INSTALL_COMMAND to the clipboard.") {
                    CopyPasteManager.getInstance().setContents(StringSelection(INSTALL_COMMAND))
                },
            ),
        )
    }
}
