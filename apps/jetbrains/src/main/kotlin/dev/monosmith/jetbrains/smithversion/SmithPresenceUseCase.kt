package dev.monosmith.jetbrains.smithversion

import com.intellij.openapi.ide.CopyPasteManager
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHVERSION_COPY_COMMAND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHVERSION_COPY_TO_THE_CLIPBOARD
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHVERSION_MONOSMITH_NEEDS_SMITH_TO_LIST
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHVERSION_SMITH_ANSWERED_WITH_AN_ERROR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SMITHVERSION_SMITH_IS_NOT_INSTALLED_HERE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
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
                onSuccess = {
                    if (it.ok) {
                        null
                    } else {
                        say(
                            JETBRAINS_SMITHVERSION_SMITH_ANSWERED_WITH_AN_ERROR,
                            it.stderr
                                .lines()
                                .firstOrNull()
                                .orEmpty()
                                .toString(),
                        )
                    }
                },
                onFailure = { say(JETBRAINS_SMITHVERSION_SMITH_IS_NOT_INSTALLED_HERE) },
            ) ?: return
        NoticeClient.warn(
            ide,
            say(JETBRAINS_SMITHVERSION_MONOSMITH_NEEDS_SMITH_TO_LIST, problem.toString(), INSTALL_COMMAND.toString()),
            listOf(
                NoticeButtonContract(
                    say(JETBRAINS_SMITHVERSION_COPY_COMMAND),
                    say(JETBRAINS_SMITHVERSION_COPY_TO_THE_CLIPBOARD, INSTALL_COMMAND.toString()),
                ) {
                    CopyPasteManager.getInstance().setContents(StringSelection(INSTALL_COMMAND))
                },
            ),
        )
    }
}
