package dev.monosmith.jetbrains.tabcompletion

import com.intellij.ide.util.PropertiesComponent
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.smithrun.SmithRunContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TABCOMPLETION_DONT_ASK_AGAIN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TABCOMPLETION_NOT_NOW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TABCOMPLETION_OFFER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_TABCOMPLETION_SET_UP
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path
import java.util.concurrent.atomic.AtomicBoolean

/**
 * The IDE side of the offer of tab completion: the answer is kept for the person (the application
 * settings of the IDE, not one project), smith is asked in the folder, and the offer is a notification.
 *
 * @property ide the IDE project
 * @property folder the folder smith runs in and the terminal starts in
 */
class TabCompletionIdeClient(
    private val ide: Project,
    private val folder: Path,
) : TabCompletionPortsContract {
    override fun remembered(): String? = PropertiesComponent.getInstance().getValue(KEY)

    override fun remember(answer: String) = PropertiesComponent.getInstance().setValue(KEY, answer)

    override fun askedThisSession(): Boolean = asked.get()

    override fun markAsked() = asked.set(true)

    override fun status(): SmithRunContract = SmithRunClient.run(null, folder, listOf("completion", "status", "--json"))

    override fun ask(
        setUp: () -> Unit,
        never: () -> Unit,
    ) = ApplicationManager.getApplication().invokeLater {
        NoticeClient.info(
            ide,
            say(JETBRAINS_TABCOMPLETION_OFFER),
            listOf(
                NoticeButtonContract(say(JETBRAINS_TABCOMPLETION_SET_UP), action = setUp),
                NoticeButtonContract(say(JETBRAINS_TABCOMPLETION_NOT_NOW), action = {}),
                NoticeButtonContract(say(JETBRAINS_TABCOMPLETION_DONT_ASK_AGAIN), action = never),
            ),
        )
    }

    override fun openTerminal() = SmithTerminalClient.run(ide, folder.toString(), SmithTerminalClient.INSTALL_COMMAND)

    private companion object {
        /** The key of the answer in the application settings of the IDE. */
        const val KEY = "monosmith.tabCompletionOffer"

        /** Whether the offer was shown since the IDE started. */
        val asked = AtomicBoolean(false)
    }
}
