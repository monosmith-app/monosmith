package dev.monosmith.jetbrains.addproject

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.addchoices.ChoicesJsonMapper
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.installdependency.InstallDependencyUseCase
import dev.monosmith.jetbrains.installtools.InstallToolsUseCase
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.OutputConsoleStore
import dev.monosmith.jetbrains.smithrun.SmithRunContract
import dev.monosmith.jetbrains.undochange.UndoChangeUseCase
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import java.nio.file.Path

/**
 * What the buttons of the notices of an add can start. The IDE gives the real flows; a test gives fakes.
 *
 * @property commit commits the change of an operation, after its own question
 */
data class AddFollowUpsContract(
    val undo: () -> Unit,
    val commit: (operation: String) -> Unit,
    val useInApp: () -> Unit,
    val installTools: () -> Unit,
    val openHealth: () -> Unit,
    val openPage: (url: String) -> Unit,
)

/**
 * Adds a project: reads what smith can add (`kinds`, `choices`, `compose --json`), asks the person
 * outcome first, shows the plan of `smith add --dry-run`, and only on a yes runs the real add with
 * `--yes`. smith commits only when the unticked box is ticked.
 */
object AddProjectUseCase {
    private const val TITLE = "MonoSmith: Add project"
    private const val COMMITTED = "Committed as one change: Undo reverts it."
    private const val UNCOMMITTED = "Not committed: review it with the Commit window, or commit it now."

    /**
     * Starts the flow in a background task.
     *
     * @param onCompose runs when the person picks "Compose a stack…" instead
     */
    fun start(
        ide: Project,
        root: Path,
        onCompose: () -> Unit,
    ) = BackgroundFlowClient.start(ide, TITLE) {
        val siblings = ListProjectsUseCase.list(root).getOrDefault(emptyList()).map { AddQuestionsUseCase.Sibling(it.name, it.language) }
        val follow =
            AddFollowUpsContract(
                undo = { UndoChangeUseCase.start(ide, root) },
                commit = { CommitChangesUseCase.start(ide, root, it) },
                useInApp = { InstallDependencyUseCase.start(ide, root, only = "workspace") },
                installTools = { InstallToolsUseCase.start(ide, root) },
                openHealth = { OutputConsoleStore.reveal(ide, "Health") },
                openPage = { BrowserUtil.browse(it) },
            )
        run(IdePortsClient.of(ide, root), siblings, follow, onCompose)
    }

    /**
     * Runs the whole flow with the given ports.
     *
     * @param siblings the projects the workspace has
     * @param onCompose runs when the person picks "Compose a stack…" instead
     * @return true when smith added the project
     */
    fun run(
        ports: FlowPortsContract,
        siblings: List<AddQuestionsUseCase.Sibling>,
        follow: AddFollowUpsContract,
        onCompose: () -> Unit,
    ): Boolean {
        val options = readOptions(ports) ?: return false
        val answer =
            AddQuestionsUseCase.ask(ports.prompt, options, siblings) { CheckPicksUseCase.check(ports.smith, options, it) }
                ?: return false
        if (answer is AddAnswerContract.Compose) {
            onCompose()
            return false
        }
        val (request, words) = answer as AddAnswerContract.Request
        val plan = ports.smith.run(AddArgumentsAlgorithm.arguments(request, apply = false))
        if (!plan.ok) {
            ports.notice.warn("Nothing was added. ${FailureTextMapper.describe(plan)}")
            return false
        }
        val follows = AddArgumentsAlgorithm.followsSiblings(plan.stdout)
        val header = "$words Nothing was changed yet. By default smith leaves the files uncommitted: review them with the Commit window."
        val answered =
            ports.prompt.plan(
                TITLE,
                header,
                withoutDryRunLine(plan.stdout),
                "Add project",
                offerCommit = true,
                checks =
                    if (follows) {
                        listOf(
                            "Skip the shape of the sibling projects and use the stack defaults (--fresh)",
                        )
                    } else {
                        emptyList()
                    },
            ) ?: return false
        val fresh = request.fresh || answered.checks.firstOrNull() == true
        val done =
            ports.smith.run(
                AddArgumentsAlgorithm.arguments(request, apply = true, fresh = fresh, commit = answered.commit),
                show = true,
            )
        ports.changed()
        // smith prints "Not committed" when it could not commit (other changes in the work tree): nothing to undo.
        val committed = answered.commit && done.stdout.lines().none { it.startsWith("Not committed") }
        return tell(ports, follow, request, committed, done)
    }

    /**
     * The plan without its closing "Dry run" line: the dialog already says that nothing was changed.
     *
     * @param text what `smith add --dry-run` printed
     * @return the plan, trimmed
     */
    internal fun withoutDryRunLine(text: String): String =
        text
            .lines()
            .filterNot { it.trim() == "Dry run: nothing was changed." }
            .joinToString("\n")
            .trim()

    private fun readOptions(ports: FlowPortsContract): AddOptionsContract? {
        val kinds = ports.smith.run(listOf("kinds", "--json"))
        val choices = ports.smith.run(listOf("choices", "--json"))
        val catalogue = ports.smith.run(listOf("compose", "--json"))
        val failed = listOf(kinds, choices, catalogue).firstOrNull { !it.ok }
        if (failed != null) {
            ports.notice.warn("Nothing was added. ${FailureTextMapper.describe(failed)}")
            return null
        }
        return runCatching {
            val kindList = OutcomesJsonMapper.toKinds(kinds.stdout)
            val found = OutcomesJsonMapper.toCatalogue(catalogue.stdout)
            AddOptionsContract(
                kindList,
                ChoicesJsonMapper.toChoices(choices.stdout),
                OutcomesJsonMapper.fitOutcomes(found.outcomes, kindList),
                found.languages,
                found.layers,
                found.about,
            )
        }.getOrElse {
            ports.notice.warn(
                "Nothing was added. smith printed its lists in a form this plugin does not read. Update the plugin, or run smith add in a terminal.",
            )
            null
        }
    }

    private fun tell(
        ports: FlowPortsContract,
        follow: AddFollowUpsContract,
        request: AddRequestContract,
        committed: Boolean,
        done: SmithRunContract,
    ): Boolean {
        val operation = "add ${request.role} ${request.name}"
        if (done.ok) {
            val buttons =
                listOfNotNull(
                    if (committed) NoticeButtonContract("Undo", "Revert the commit of this add.") { follow.undo() } else null,
                    if (committed) {
                        null
                    } else {
                        NoticeButtonContract(
                            "Commit now",
                            "Make one commit that Undo can revert; you choose the files first.",
                        ) {
                            follow.commit(operation)
                        }
                    },
                    if (request.role ==
                        "lib"
                    ) {
                        NoticeButtonContract("Use in an app…", "Add this library to an app.") { follow.useInApp() }
                    } else {
                        null
                    },
                )
            ports.notice.info(
                "${request.name} is added and the CI pipelines are regenerated. ${if (committed) COMMITTED else UNCOMMITTED}",
                buttons,
            )
            NeededToolsMapper.parse(done.stdout)?.let { tools ->
                ports.notice.warn(
                    "${request.name} needs tools this machine lacks. ${tools.problems.joinToString(" ")} " +
                        "Without them the first doctor run warns. Install tools asks how to get each one; nothing is installed before you choose.",
                    listOf(
                        NoticeButtonContract(
                            "Install tools",
                            "Choose how to get ${tools.names.joinToString(", ")} with smith tools install.",
                        ) { follow.installTools() },
                    ),
                )
            }
            return true
        }
        val everything = "${done.stdout}\n${done.stderr}"
        val failure = AddFailureMapper.doctorFailure(everything)
        if (failure != null) {
            tellFailing(ports, follow, operation, failure)
            return false
        }
        val tool = AddFailureMapper.missingTool(everything)
        if (tool != null) {
            tellUnchecked(ports, follow, request.name, tool)
            return false
        }
        ports.notice.warn("Nothing was added. ${FailureTextMapper.describe(done)}")
        return false
    }

    private fun tellFailing(
        ports: FlowPortsContract,
        follow: AddFollowUpsContract,
        operation: String,
        failure: DoctorFailureContract,
    ) = ports.notice.warn(
        AddFailureMapper.describe(failure),
        listOf(
            NoticeButtonContract("Install tools", "Install the missing tools with smith tools install.") { follow.installTools() },
            NoticeButtonContract("Open Health", "Open the Health tab of the MonoSmith tool window.") { follow.openHealth() },
            NoticeButtonContract("Commit anyway", "Commit the project although it fails its checks; you choose the files first.") {
                follow.commit(operation)
            },
            NoticeButtonContract("Keep uncommitted", "Leave the files in your work tree.") {},
        ),
    )

    private fun tellUnchecked(
        ports: FlowPortsContract,
        follow: AddFollowUpsContract,
        name: String,
        tool: MissingToolContract,
    ) {
        val where = tool.dir?.let { " in $it" }.orEmpty()
        val files = if (tool.files > 0) " (${tool.files} files)" else ""
        val check = tool.check
        val buttons =
            listOfNotNull(
                tool.url?.let { url ->
                    NoticeButtonContract("Install page", "Open the page that tells how to get ${tool.title}.") { follow.openPage(url) }
                },
                NoticeButtonContract("Install tools", "Install the missing tools with smith tools install.") { follow.installTools() },
                if (check ==
                    null
                ) {
                    null
                } else {
                    NoticeButtonContract(
                        "Check again",
                        "Run the first check of the new project again.",
                    ) { ports.smith.run(check, show = true) }
                },
                NoticeButtonContract("Skip the check", "Continue without the check. The Health tab keeps reminding you.") {},
            )
        ports.notice.warn(
            "$name was created$where$files, and could not be checked yet: ${tool.title} is not installed. Nothing is lost: your project stays on disk.",
            buttons,
        )
    }
}
