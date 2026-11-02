package dev.monosmith.jetbrains.workspacetools

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.PickProjectUseCase
import java.nio.file.Path

/** Small workspace-wide actions that each end in one smith command: the launcher, the agent files, the divergences, the environments of a web app and the explanations. */
object WorkspaceToolsUseCase {
    /** Writes the launcher that makes a laptop and CI run the same smith (`smith bootstrap`), after its plan. */
    fun writeLauncher(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: write the launcher") {
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = "MonoSmith: write the smith launcher?",
                planArgs = listOf("bootstrap", "--dry-run"),
                header = "The files are meant to be committed, and one that exists is replaced. Nothing is written until you answer.",
                okText = "Write the launcher",
                offerCommit = true,
                apply = { listOf("bootstrap", "--yes") },
                done = "The launcher is written",
                operation = "bootstrap",
            ),
        )
    }

    /**
     * Installs the MonoSmith agent files and MCP server for Claude Code. smith prints the diffs first
     * and never overwrites a file the team edited. Nothing is committed.
     */
    fun installAgent(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: install the agent") {
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = "MonoSmith: install the MCP server and agent?",
                planArgs = listOf("agent", "install", "--claude"),
                header = "smith writes .mcp.json and .claude/agents/smith.md for Claude Code. Nothing is written until you answer.",
                okText = "Install",
                offerCommit = false,
                apply = { listOf("agent", "install", "--claude", "--yes") },
                done = "The agent files are written",
            ),
        )
    }

    /** Shows the divergences from the defaults (override, ejected, drift) in the Output tab. */
    fun divergences(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: divergences") { SmithRunClient.run(ide, root, listOf("divergences"), show = true) }

    /** Lists the environments of the web apps and the task that builds each. */
    fun listEnvironments(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: environments") {
        SmithRunClient.run(ide, root, listOf("environments", "list"), show = true)
    }

    /**
     * Adds or removes environments of a web app: asks the app and the names, shows the plan, then runs it.
     *
     * @param remove true removes the environments (their .env files stay); false adds them
     * @param project the web app; null asks which
     */
    fun changeEnvironments(
        ide: Project,
        root: Path,
        remove: Boolean,
        project: String?,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: environments") {
        val verb = if (remove) "remove" else "add"
        val nodes = ListProjectsUseCase.list(root).getOrElse { return@start NoticeClient.warn(ide, "Nothing was changed. ${it.message}") }
        val app =
            if (project != null) {
                nodes.firstOrNull { it.name == project } ?: return@start
            } else {
                PickProjectUseCase.pick(ide, nodes.filter { it.kind == "app" && it.language == "node" }, "$verb environments of a web app")
                    ?: return@start
            }
        val names =
            PromptClient
                .text(
                    ide,
                    "MonoSmith: $verb environments of ${app.name}",
                    "The environments to $verb (lowercase words, separated by spaces or commas):",
                ) {
                    if (it.isEmpty()) "Name at least one environment, for example qa." else null
                }?.split(Regex("[\\s,;]+"))
                ?.filter { it.isNotEmpty() } ?: return@start
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = "MonoSmith: $verb ${names.joinToString(", ")} on ${app.name}?",
                planArgs = listOf("environments", verb, app.name) + names + "--dry-run",
                header =
                    if (remove) {
                        "The files, scripts and smith.toml lines it changes. The .env files, which hold your values, stay."
                    } else {
                        "The files it creates, the scripts it changes and the line in smith.toml."
                    },
                okText = verb.replaceFirstChar { it.uppercase() },
                offerCommit = true,
                apply = { listOf("environments", verb, app.name) + names + "--yes" },
                done = "The environments of ${app.name} changed",
                operation = "environments $verb ${app.name}",
            ),
        )
    }

    /**
     * Says why a file, a project, a task or a setting is as it is (`smith why`), or how a project's tasks
     * resolved (`smith explain`), in the Output tab.
     *
     * @param target a file relative to the workspace, or null to ask for a file, project, task or setting path
     */
    fun why(
        ide: Project,
        root: Path,
        target: String?,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: why") {
        val what =
            target ?: PromptClient.text(ide, "MonoSmith: why is this here?", "A file, a folder, a project, a task or a setting path:") {
                if (it.isEmpty()) "Type what you want explained." else null
            } ?: return@start
        SmithRunClient.run(ide, root, listOf("why", what), show = true)
    }

    /** Shows how the tasks of a project resolved and where each field came from (`smith explain <project>`). */
    fun explain(
        ide: Project,
        root: Path,
        project: String?,
    ) = BackgroundFlowClient.start(ide, "MonoSmith: explain a project") {
        val nodes = ListProjectsUseCase.list(root).getOrElse { return@start NoticeClient.warn(ide, "Nothing was explained. ${it.message}") }
        val name = project ?: PickProjectUseCase.pick(ide, nodes, "explain a project")?.name ?: return@start
        SmithRunClient.run(ide, root, listOf("explain", name), show = true)
    }
}
