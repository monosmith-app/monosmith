package dev.monosmith.jetbrains.devservers

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.taskrun.RunSmithCommandUseCase
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
                "Runs its start task.",
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
    ) = BackgroundFlowClient.start(ide, "MonoSmith: dev servers") {
        val projects =
            ListProjectsUseCase
                .list(
                    root,
                ).getOrElse { return@start NoticeClient.warn(ide, "No server was started. ${it.message}") }
        val rows = rows(projects)
        if (rows.isEmpty()) return@start NoticeClient.info(ide, "No project has a start task.")
        val names =
            PromptClient.pickMany(
                ide,
                "MonoSmith: dev servers",
                "Tick the projects to start. Each runs in the Run window; its Stop button stops it.",
                rows,
                "Start",
            )
                ?: return@start
        if (names.isEmpty()) return@start
        PromptClient.onUi { RunSmithCommandUseCase.run(ide, root, "smith dev ${names.joinToString(" ")}") { arguments(it, names) } }
    }
}
