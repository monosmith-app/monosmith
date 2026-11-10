package dev.monosmith.jetbrains.devservers

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
import dev.monosmith.jetbrains.words.SmithWordsConfig.COMMANDS_GROUP_START
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEVSERVERS_MONOSMITH_DEV_SERVERS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEVSERVERS_NO_PROJECT_HAS_A_START
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEVSERVERS_NO_SERVER_WAS_STARTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEVSERVERS_RUNS_ITS_START_TASK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEVSERVERS_SMITH_DEV
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_DEVSERVERS_TICK_THE_PROJECTS_TO_START
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.SmithProjectContract
import java.nio.file.Path

/**
 * Starts the dev servers of the projects that have a `start` task (`smith dev <projects>`) in a Run window.
 * The Run window has the Stop button, so a server is stopped the way every IDE process is. Nothing is
 * changed in the workspace.
 */
object DevServersUseCase {
    /** The projects that have a start task, as the lines of the question; none is ticked at first. */
    fun rows(projects: List<SmithProjectContract>): List<ChoiceRowContract> =
        projects.filter { "start" in it.tasks }.map {
            ChoiceRowContract(
                it.name,
                it.name,
                "${it.kind} · ${it.language}",
                say(JETBRAINS_DEVSERVERS_RUNS_ITS_START_TASK),
            )
        }

    /** `smith dev <projects>`; the names are separate arguments. */
    fun arguments(
        smith: String,
        names: List<String>,
    ): List<String> = listOf(smith, "dev") + names

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_DEVSERVERS_MONOSMITH_DEV_SERVERS)) {
        val projects =
            ListProjectsUseCase
                .list(
                    root,
                ).getOrElse { return@start NoticeClient.warn(ide, say(JETBRAINS_DEVSERVERS_NO_SERVER_WAS_STARTED, it.message.toString())) }
        val rows = rows(projects)
        if (rows.isEmpty()) return@start NoticeClient.info(ide, say(JETBRAINS_DEVSERVERS_NO_PROJECT_HAS_A_START))
        val names =
            PromptClient.pickMany(
                ide,
                say(JETBRAINS_DEVSERVERS_MONOSMITH_DEV_SERVERS),
                say(JETBRAINS_DEVSERVERS_TICK_THE_PROJECTS_TO_START),
                rows,
                say(COMMANDS_GROUP_START),
            )
                ?: return@start
        if (names.isEmpty()) return@start
        PromptClient.onUi {
            RunSmithCommandUseCase.run(
                ide,
                root,
                say(JETBRAINS_DEVSERVERS_SMITH_DEV, names.joinToString(" ").toString()),
            ) { arguments(it, names) }
        }
    }
}
