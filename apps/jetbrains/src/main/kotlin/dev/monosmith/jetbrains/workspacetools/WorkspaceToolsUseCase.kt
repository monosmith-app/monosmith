package dev.monosmith.jetbrains.workspacetools

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.words.SmithWordsConfig.COMMANDS_DESCRIBE_BOOTSTRAP_TITLE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_NOTHING_WAS_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_A_FILE_A_FOLDER_A
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_ENVIRONMENTS_OF_A_WEB_APP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_EXPLAIN_A_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_DIVERGENCES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_ENVIRONMENTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_ENVIRONMENTS_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_EXPLAIN_A_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_INSTALL_THE_AGENT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_INSTALL_THE_MCP_SERVER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_ON
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_WHY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_WHY_IS_THIS_HERE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_WRITE_THE_LAUNCHER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_MONOSMITH_WRITE_THE_SMITH_LAUNCHER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_NAME_AT_LEAST_ONE_ENVIRONMENT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_NOTHING_WAS_EXPLAINED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_SMITH_WRITES_MCP_JSON_AND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_THE_AGENT_FILES_ARE_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_THE_ENVIRONMENTS_OF_CHANGED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_THE_ENVIRONMENTS_TO_LOWERCASE_WORDS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_THE_FILES_ARE_MEANT_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_THE_FILES_IT_CREATES_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_THE_FILES_SCRIPTS_AND_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_THE_LAUNCHER_IS_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_WORKSPACETOOLS_TYPE_WHAT_YOU_WANT_EXPLAINED
import dev.monosmith.jetbrains.words.SmithWordsConfig.SETTINGSVIEW_GROUP_INSTALL
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.PickProjectUseCase
import java.nio.file.Path

/** Small workspace-wide actions that each end in one smith command: the launcher, the agent files, the divergences, the environments of a web app and the explanations. */
object WorkspaceToolsUseCase {
    /** Writes the launcher that makes a laptop and CI run the same smith (`smith bootstrap`), after its plan. */
    fun writeLauncher(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_WRITE_THE_LAUNCHER)) {
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_WRITE_THE_SMITH_LAUNCHER),
                planArgs = listOf("bootstrap", "--dry-run"),
                header = say(JETBRAINS_WORKSPACETOOLS_THE_FILES_ARE_MEANT_TO),
                okText = say(COMMANDS_DESCRIBE_BOOTSTRAP_TITLE),
                offerCommit = true,
                apply = { listOf("bootstrap", "--yes") },
                done = say(JETBRAINS_WORKSPACETOOLS_THE_LAUNCHER_IS_WRITTEN),
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_INSTALL_THE_AGENT)) {
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title = say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_INSTALL_THE_MCP_SERVER),
                planArgs = listOf("agent", "install", "--claude"),
                header = say(JETBRAINS_WORKSPACETOOLS_SMITH_WRITES_MCP_JSON_AND),
                okText = say(SETTINGSVIEW_GROUP_INSTALL),
                offerCommit = false,
                apply = { listOf("agent", "install", "--claude", "--yes") },
                done = say(JETBRAINS_WORKSPACETOOLS_THE_AGENT_FILES_ARE_WRITTEN),
            ),
        )
    }

    /** Shows the divergences from the defaults (override, ejected, drift) in the Output tab. */
    fun divergences(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(
        ide,
        say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_DIVERGENCES),
    ) { SmithRunClient.run(ide, root, listOf("divergences"), show = true) }

    /** Lists the environments of the web apps and the task that builds each. */
    fun listEnvironments(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_ENVIRONMENTS)) {
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_ENVIRONMENTS)) {
        val verb = if (remove) "remove" else "add"
        val nodes =
            ListProjectsUseCase
                .list(
                    root,
                ).getOrElse { return@start NoticeClient.warn(ide, say(JETBRAINS_SWAPCHOICE_NOTHING_WAS_CHANGED, it.message.toString())) }
        val app =
            if (project != null) {
                nodes.firstOrNull { it.name == project } ?: return@start
            } else {
                PickProjectUseCase.pick(
                    ide,
                    nodes.filter { it.kind == "app" && it.language == "node" },
                    say(JETBRAINS_WORKSPACETOOLS_ENVIRONMENTS_OF_A_WEB_APP, verb.toString()),
                )
                    ?: return@start
            }
        val names =
            PromptClient
                .text(
                    ide,
                    say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_ENVIRONMENTS_OF, verb.toString(), app.name.toString()),
                    say(JETBRAINS_WORKSPACETOOLS_THE_ENVIRONMENTS_TO_LOWERCASE_WORDS, verb.toString()),
                ) {
                    if (it.isEmpty()) say(JETBRAINS_WORKSPACETOOLS_NAME_AT_LEAST_ONE_ENVIRONMENT) else null
                }?.split(Regex("[\\s,;]+"))
                ?.filter { it.isNotEmpty() } ?: return@start
        PlanThenApplyUseCase.run(
            ide,
            root,
            PlanStepContract(
                title =
                    say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_ON, verb.toString(), names.joinToString(", ").toString(), app.name.toString()),
                planArgs = listOf("environments", verb, app.name) + names + "--dry-run",
                header =
                    if (remove) {
                        say(JETBRAINS_WORKSPACETOOLS_THE_FILES_SCRIPTS_AND_SMITH)
                    } else {
                        say(JETBRAINS_WORKSPACETOOLS_THE_FILES_IT_CREATES_THE)
                    },
                okText = verb.replaceFirstChar { it.uppercase() },
                offerCommit = true,
                apply = { listOf("environments", verb, app.name) + names + "--yes" },
                done = say(JETBRAINS_WORKSPACETOOLS_THE_ENVIRONMENTS_OF_CHANGED, app.name.toString()),
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
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_WHY)) {
        val what =
            target
                ?: PromptClient.text(
                    ide,
                    say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_WHY_IS_THIS_HERE),
                    say(JETBRAINS_WORKSPACETOOLS_A_FILE_A_FOLDER_A),
                ) {
                    if (it.isEmpty()) say(JETBRAINS_WORKSPACETOOLS_TYPE_WHAT_YOU_WANT_EXPLAINED) else null
                } ?: return@start
        SmithRunClient.run(ide, root, listOf("why", what), show = true)
    }

    /** Shows how the tasks of a project resolved and where each field came from (`smith explain <project>`). */
    fun explain(
        ide: Project,
        root: Path,
        project: String?,
    ) = BackgroundFlowClient.start(ide, say(JETBRAINS_WORKSPACETOOLS_MONOSMITH_EXPLAIN_A_PROJECT)) {
        val nodes =
            ListProjectsUseCase
                .list(
                    root,
                ).getOrElse {
                    return@start NoticeClient.warn(
                        ide,
                        say(JETBRAINS_WORKSPACETOOLS_NOTHING_WAS_EXPLAINED, it.message.toString()),
                    )
                }
        val name = project ?: PickProjectUseCase.pick(ide, nodes, say(JETBRAINS_WORKSPACETOOLS_EXPLAIN_A_PROJECT))?.name ?: return@start
        SmithRunClient.run(ide, root, listOf("explain", name), show = true)
    }
}
