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
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_ADD_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_ADD_THIS_LIBRARY_TO_AN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_CHOOSE_HOW_TO_GET_WITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COMMITTED_AS_ONE_CHANGE_UNDO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COMMIT_ANYWAY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COMMIT_NOW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COMMIT_THE_PROJECT_ALTHOUGH_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_CONTINUE_WITHOUT_THE_CHECK_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_INSTALL_PAGE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_INSTALL_THE_MISSING_TOOLS_WITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_INSTALL_TOOLS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_IS_ADDED_AND_THE_CI
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_KEEP_UNCOMMITTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_LEAVE_THE_FILES_IN_YOUR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_MAKE_ONE_COMMIT_THAT_UNDO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_MONOSMITH_ADD_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NEEDS_TOOLS_THIS_MACHINE_LACKS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED_SMITH_PRINTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NOTHING_WAS_CHANGED_YET_BY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NOT_COMMITTED_REVIEW_IT_WITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_OPEN_HEALTH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_OPEN_THE_HEALTH_TAB_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_OPEN_THE_PAGE_THAT_TELLS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_REVERT_THE_COMMIT_OF_THIS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_RUN_THE_FIRST_CHECK_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_SKIP_THE_CHECK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_SKIP_THE_SHAPE_OF_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_UNDO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_USE_IN_AN_APP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_WAS_CREATED_AND_COULD_NOT
import dev.monosmith.jetbrains.words.SmithWordsConfig.RESULTDIALOG_CHECK_LABEL
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
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
    private val TITLE = say(JETBRAINS_ADDPROJECT_MONOSMITH_ADD_PROJECT)
    private val COMMITTED = say(JETBRAINS_ADDPROJECT_COMMITTED_AS_ONE_CHANGE_UNDO)
    private val UNCOMMITTED = say(JETBRAINS_ADDPROJECT_NOT_COMMITTED_REVIEW_IT_WITH)

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
            ports.notice.warn(say(JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED, FailureTextMapper.describe(plan).toString()))
            return false
        }
        val follows = AddArgumentsAlgorithm.followsSiblings(plan.stdout)
        val header = say(JETBRAINS_ADDPROJECT_NOTHING_WAS_CHANGED_YET_BY, words.toString())
        val answered =
            ports.prompt.plan(
                TITLE,
                header,
                withoutDryRunLine(plan.stdout),
                say(JETBRAINS_ADDPROJECT_ADD_PROJECT),
                offerCommit = true,
                checks =
                    if (follows) {
                        listOf(
                            say(JETBRAINS_ADDPROJECT_SKIP_THE_SHAPE_OF_THE),
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
            ports.notice.warn(say(JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED, FailureTextMapper.describe(failed).toString()))
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
                say(JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED_SMITH_PRINTED),
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
                    if (committed) {
                        NoticeButtonContract(
                            say(JETBRAINS_ADDPROJECT_UNDO),
                            say(JETBRAINS_ADDPROJECT_REVERT_THE_COMMIT_OF_THIS),
                        ) {
                            follow.undo()
                        }
                    } else {
                        null
                    },
                    if (committed) {
                        null
                    } else {
                        NoticeButtonContract(
                            say(JETBRAINS_ADDPROJECT_COMMIT_NOW),
                            say(JETBRAINS_ADDPROJECT_MAKE_ONE_COMMIT_THAT_UNDO),
                        ) {
                            follow.commit(operation)
                        }
                    },
                    if (request.role ==
                        "lib"
                    ) {
                        NoticeButtonContract(
                            say(JETBRAINS_ADDPROJECT_USE_IN_AN_APP),
                            say(JETBRAINS_ADDPROJECT_ADD_THIS_LIBRARY_TO_AN),
                        ) { follow.useInApp() }
                    } else {
                        null
                    },
                )
            ports.notice.info(
                say(
                    JETBRAINS_ADDPROJECT_IS_ADDED_AND_THE_CI,
                    request.name.toString(),
                    (if (committed) COMMITTED else UNCOMMITTED).toString(),
                ),
                buttons,
            )
            NeededToolsMapper.parse(done.stdout)?.let { tools ->
                ports.notice.warn(
                    say(
                        JETBRAINS_ADDPROJECT_NEEDS_TOOLS_THIS_MACHINE_LACKS,
                        request.name.toString(),
                        tools.problems.joinToString(" ").toString(),
                    ),
                    listOf(
                        NoticeButtonContract(
                            say(JETBRAINS_ADDPROJECT_INSTALL_TOOLS),
                            say(JETBRAINS_ADDPROJECT_CHOOSE_HOW_TO_GET_WITH, tools.names.joinToString(", ").toString()),
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
        ports.notice.warn(say(JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED, FailureTextMapper.describe(done).toString()))
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
            NoticeButtonContract(
                say(JETBRAINS_ADDPROJECT_INSTALL_TOOLS),
                say(JETBRAINS_ADDPROJECT_INSTALL_THE_MISSING_TOOLS_WITH),
            ) { follow.installTools() },
            NoticeButtonContract(
                say(JETBRAINS_ADDPROJECT_OPEN_HEALTH),
                say(JETBRAINS_ADDPROJECT_OPEN_THE_HEALTH_TAB_OF),
            ) { follow.openHealth() },
            NoticeButtonContract(say(JETBRAINS_ADDPROJECT_COMMIT_ANYWAY), say(JETBRAINS_ADDPROJECT_COMMIT_THE_PROJECT_ALTHOUGH_IT)) {
                follow.commit(operation)
            },
            NoticeButtonContract(say(JETBRAINS_ADDPROJECT_KEEP_UNCOMMITTED), say(JETBRAINS_ADDPROJECT_LEAVE_THE_FILES_IN_YOUR)) {},
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
                    NoticeButtonContract(
                        say(JETBRAINS_ADDPROJECT_INSTALL_PAGE),
                        say(JETBRAINS_ADDPROJECT_OPEN_THE_PAGE_THAT_TELLS, tool.title.toString()),
                    ) { follow.openPage(url) }
                },
                NoticeButtonContract(
                    say(JETBRAINS_ADDPROJECT_INSTALL_TOOLS),
                    say(JETBRAINS_ADDPROJECT_INSTALL_THE_MISSING_TOOLS_WITH),
                ) { follow.installTools() },
                if (check ==
                    null
                ) {
                    null
                } else {
                    NoticeButtonContract(
                        say(RESULTDIALOG_CHECK_LABEL),
                        say(JETBRAINS_ADDPROJECT_RUN_THE_FIRST_CHECK_OF),
                    ) { ports.smith.run(check, show = true) }
                },
                NoticeButtonContract(say(JETBRAINS_ADDPROJECT_SKIP_THE_CHECK), say(JETBRAINS_ADDPROJECT_CONTINUE_WITHOUT_THE_CHECK_THE)) {},
            )
        ports.notice.warn(
            say(JETBRAINS_ADDPROJECT_WAS_CREATED_AND_COULD_NOT, name.toString(), where.toString(), files.toString(), tool.title.toString()),
            buttons,
        )
    }
}
