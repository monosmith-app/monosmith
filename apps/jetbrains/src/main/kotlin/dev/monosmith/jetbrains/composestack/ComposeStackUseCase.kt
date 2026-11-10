package dev.monosmith.jetbrains.composestack

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.addproject.ProjectNameValidator
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.installtools.InstallToolsUseCase
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_SUGGESTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COMMITTED_AS_ONE_CHANGE_UNDO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_COMMIT_NOW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_MAKE_ONE_COMMIT_THAT_UNDO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_NOT_COMMITTED_REVIEW_IT_WITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_2_OF_3_NAME
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_3_OF_3_PLAN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_ADD_THE_PROJECT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_CONFLICT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_INSTALL_THE_TOOLS_FIRST
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_IS_ADDED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_IS_NOT_PART_OF_ONE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_KEEP_IT_AS_A_RECIPE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_MONOSMITH_COMPOSE_A_STACK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_NEEDS_AN_ANSWER_ON_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_NEXT_CHECK
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_NOTHING_IS_WRITTEN_YET_BY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_NOTHING_WAS_ADDED_THE_COMMAND
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_NOTHING_WAS_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_SMITH_DID_NOT_CHECK_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_SMITH_MADE_NO_PLAN_FOR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_SMITH_PRINTED_ITS_CATALOGUE_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_STILL_NEEDS_A_CHOICE_ON
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_THE_NAME_OF_THE_NEW
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_THE_RECIPE_IS_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_COMPOSESTACK_TICK_AT_LEAST_ONE_CHOICE
import dev.monosmith.jetbrains.workspace.PortProjectsUseCase
import java.nio.file.Path

/**
 * Composes a stack: shows the catalogue, checks the ticks with smith after each round (a conflict is
 * explained, an open question is asked, nothing is chosen silently), then names the project and shows
 * the plan. The plan can be run, run with a commit, or kept as a recipe. Nothing is written before the
 * person says so. The steps are numbered: 1 pick, 2 name, 3 plan.
 */
object ComposeStackUseCase {
    private val TITLE = SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_MONOSMITH_COMPOSE_A_STACK)
    private val COMMITTED = SmithWordsConfig.say(JETBRAINS_ADDPROJECT_COMMITTED_AS_ONE_CHANGE_UNDO)
    private val UNCOMMITTED = SmithWordsConfig.say(JETBRAINS_ADDPROJECT_NOT_COMMITTED_REVIEW_IT_WITH)

    /**
     * The step that shows the catalogue and returns the ids ticked, or null when the person cancels.
     * The IDE one is a dialog; a test gives a function that answers from a script.
     */
    fun interface PickStack {
        /** Shows [report] with [ticked] ticked and [message] on top; returns the ids ticked, or null on Cancel. */
        fun pick(
            report: ComposeReportContract,
            ticked: List<String>,
            message: String,
        ): List<String>?
    }

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) {
        val pick =
            PickStack { report, ticked, message ->
                PromptClient.onUi {
                    val dialog =
                        ComposePickDialogHandler(
                            ide,
                            report.layers,
                            report.items,
                            ticked,
                            message,
                            SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_NEXT_CHECK),
                        )
                    if (dialog.showAndGet()) dialog.selected else null
                }
            }
        run(IdePortsClient.of(ide, root), pick, { InstallToolsUseCase.start(ide, root) }) { CommitChangesUseCase.start(ide, root, it) }
    }

    /**
     * Runs the flow with the given ports.
     *
     * @param pick shows the catalogue and returns the ticks
     * @param installTools starts the install of the missing tools; the plan offers it
     * @param commitLater starts the Commit action for an operation
     * @return true when smith added the project or wrote the recipe
     */
    fun run(
        ports: FlowPortsContract,
        pick: PickStack,
        installTools: () -> Unit,
        commitLater: (String) -> Unit,
    ): Boolean = compose(ports, pick, installTools, commitLater)

    private fun read(
        ports: FlowPortsContract,
        args: List<String>,
    ): ComposeReportContract? {
        val run = ports.smith.run(args)
        if (!run.ok) {
            ports.notice.warn(SmithWordsConfig.say(JETBRAINS_ADDPROJECT_NOTHING_WAS_ADDED, FailureTextMapper.describe(run).toString()))
            return null
        }
        return runCatching { ComposeReportJsonMapper.toReport(run.stdout) }.getOrElse {
            ports.notice.warn(
                SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_SMITH_PRINTED_ITS_CATALOGUE_IN),
            )
            null
        }
    }

    private fun compose(
        ports: FlowPortsContract,
        pick: PickStack,
        installTools: () -> Unit,
        commitLater: (String) -> Unit,
    ): Boolean {
        val first = read(ports, ComposeStackAlgorithm.arguments(null, emptyList())) ?: return false
        var ticked = emptyList<String>()
        var message = ""
        while (true) {
            val chosen = pick.pick(first, ticked, message)
            if (chosen.isNullOrEmpty()) return false
            val (settled, note) = ComposeStackAlgorithm.settleRivals(first.items, ticked, chosen)
            ticked = settled

            fun say(text: String) = listOf(note, text).filter { it.isNotEmpty() }.joinToString(" ")
            val ticks = ticked.mapNotNull { id -> first.items.firstOrNull { it.id == id } }
            val language =
                ticks.firstOrNull { it.kind == "language" }?.name ?: ticks.firstOrNull { it.kind == "choice" }?.language.orEmpty()
            val plugins = ticks.filter { it.kind == "plugin" }
            plugins.forEach {
                ports.notice.info(SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_IS_NOT_PART_OF_ONE, it.title.toString(), it.route.toString()))
            }
            ticked = ticked.filter { id -> plugins.none { it.id == id } }
            val picks = ticks.filter { it.kind == "choice" }.map { it.name }
            if (picks.isEmpty()) {
                message = say(SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_TICK_AT_LEAST_ONE_CHOICE))
                continue
            }
            val checked = read(ports, ComposeStackAlgorithm.arguments(language, picks)) ?: return false
            val check = checked.check
            if (check == null) {
                ports.notice.warn(SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_SMITH_DID_NOT_CHECK_THE))
                return false
            }
            if (check.conflict.isNotEmpty()) {
                message = say(SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_CONFLICT, check.conflict.toString()))
                continue
            }
            val question = check.question
            if (question != null) {
                val rows =
                    question.options
                        .sortedByDescending {
                            it.suggested
                        }.map {
                            ChoiceRowContract(
                                it.name,
                                it.title,
                                if (it.suggested) SmithWordsConfig.say(JETBRAINS_ADDCHOICES_SUGGESTED) else "",
                                it.description,
                            )
                        }
                val answer =
                    ports.prompt.pickOne(
                        "$TITLE: ${question.axis}",
                        SmithWordsConfig.say(
                            JETBRAINS_COMPOSESTACK_NEEDS_AN_ANSWER_ON_THE,
                            question.choice.toString(),
                            question.axis.toString(),
                            question.about.toString(),
                        ),
                        rows,
                    )
                val asked =
                    answer?.let { a ->
                        first.items.firstOrNull {
                            it.kind == "choice" && it.name == a.name &&
                                it.language == check.language
                        }
                    }
                if (asked != null) ticked = ticked + asked.id
                message =
                    say(
                        if (asked ==
                            null
                        ) {
                            SmithWordsConfig.say(
                                JETBRAINS_COMPOSESTACK_STILL_NEEDS_A_CHOICE_ON,
                                question.choice.toString(),
                                question.axis.toString(),
                            )
                        } else {
                            ""
                        },
                    )
                continue
            }
            if (!check.complete) {
                message = say(check.missing)
                continue
            }
            if (note.isNotEmpty()) ports.notice.info(note)
            return finish(ports, check, ComposeStackAlgorithm.stackSummary(first.items, check.language, picks), installTools, commitLater)
        }
    }

    private fun finish(
        ports: FlowPortsContract,
        check: ComposeCheckContract,
        stackName: String,
        installTools: () -> Unit,
        commitLater: (String) -> Unit,
    ): Boolean {
        // The name comes before the plan because the files of the plan depend on it (apps/<name>/).
        val existing = PortProjectsUseCase.list(ports).getOrDefault(emptyList()).map { it.name }
        val name =
            ports.prompt.text(
                SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_2_OF_3_NAME, TITLE.toString()),
                SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_THE_NAME_OF_THE_NEW, stackName.toString()),
            ) {
                ProjectNameValidator.problem(it, existing)
            } ?: return false
        val planned =
            read(ports, ComposeStackAlgorithm.arguments(check.language, check.selection, kind = "app", name = name)) ?: return false
        val plan = planned.plan
        if (plan == null || plan.argv.size < 2) {
            ports.notice.warn(SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_SMITH_MADE_NO_PLAN_FOR))
            return false
        }
        val missing = ComposeStackAlgorithm.toolsToInstall(plan)
        val answer =
            ports.prompt.plan(
                SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_3_OF_3_PLAN, TITLE.toString()),
                SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_NOTHING_IS_WRITTEN_YET_BY),
                ComposeStackAlgorithm.describePlan(plan),
                SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_ADD_THE_PROJECT),
                offerCommit = true,
                others =
                    listOfNotNull(
                        if (missing.isNotEmpty()) SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_INSTALL_THE_TOOLS_FIRST) else null,
                        if (plan.recipeArgv.size > 1) SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_KEEP_IT_AS_A_RECIPE) else null,
                    ),
            ) ?: return false
        val others = listOfNotNull(if (missing.isNotEmpty()) "tools" else null, if (plan.recipeArgv.size > 1) "recipe" else null)
        when (others.getOrNull(answer.other)) {
            "tools" -> {
                installTools()
                return false
            }

            "recipe" -> {
                val done = ports.smith.run(plan.recipeArgv.drop(1), show = true)
                if (!done.ok) {
                    ports.notice.warn(
                        SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_NOTHING_WAS_WRITTEN, FailureTextMapper.describe(done).toString()),
                    )
                    return false
                }
                ports.notice.info(done.lines.lastOrNull() ?: SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_THE_RECIPE_IS_WRITTEN))
                return true
            }
        }
        val args = plan.argv.drop(1) + (if (answer.commit) listOf("--commit") else emptyList()) + "--yes"
        val done = ports.smith.run(args, show = true)
        ports.changed()
        if (!done.ok) {
            val problem = FailureTextMapper.describe(done)
            val command = TaskCommandAlgorithm.display(listOf("smith") + args)
            ports.notice.warn(
                SmithWordsConfig.say(JETBRAINS_COMPOSESTACK_NOTHING_WAS_ADDED_THE_COMMAND, problem.toString(), command.toString()),
            )
            return false
        }
        val buttons =
            if (answer.commit) {
                emptyList()
            } else {
                listOf(
                    NoticeButtonContract(
                        SmithWordsConfig.say(JETBRAINS_ADDPROJECT_COMMIT_NOW),
                        SmithWordsConfig.say(JETBRAINS_ADDPROJECT_MAKE_ONE_COMMIT_THAT_UNDO),
                    ) {
                        commitLater("add app $name")
                    },
                )
            }
        ports.notice.info(
            SmithWordsConfig.say(
                JETBRAINS_COMPOSESTACK_IS_ADDED,
                name.toString(),
                (if (answer.commit) COMMITTED else UNCOMMITTED).toString(),
            ),
            buttons,
        )
        return true
    }
}
