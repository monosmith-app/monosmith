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
import dev.monosmith.jetbrains.workspace.PortProjectsUseCase
import java.nio.file.Path

/**
 * Composes a stack: shows the catalogue, checks the ticks with smith after each round (a conflict is
 * explained, an open question is asked, nothing is chosen silently), then names the project and shows
 * the plan. The plan can be run, run with a commit, or kept as a recipe. Nothing is written before the
 * person says so. The steps are numbered: 1 pick, 2 name, 3 plan.
 */
object ComposeStackUseCase {
    private const val TITLE = "MonoSmith: compose a stack"
    private const val COMMITTED = "Committed as one change: Undo reverts it."
    private const val UNCOMMITTED = "Not committed: review it with the Commit window, or commit it now."

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
                    val dialog = ComposePickDialogHandler(ide, report.layers, report.items, ticked, message, "Next: check")
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
            ports.notice.warn("Nothing was added. ${FailureTextMapper.describe(run)}")
            return null
        }
        return runCatching { ComposeReportJsonMapper.toReport(run.stdout) }.getOrElse {
            ports.notice.warn(
                "smith printed its catalogue in a form this plugin does not read. Update the plugin, or run smith compose in a terminal.",
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
            plugins.forEach { ports.notice.info("${it.title} is not part of one project. Turn it on with: ${it.route}") }
            ticked = ticked.filter { id -> plugins.none { it.id == id } }
            val picks = ticks.filter { it.kind == "choice" }.map { it.name }
            if (picks.isEmpty()) {
                message = say("Tick the choice that writes the project too: a framework, a desktop shell or a cloud host.")
                continue
            }
            val checked = read(ports, ComposeStackAlgorithm.arguments(language, picks)) ?: return false
            val check = checked.check
            if (check == null) {
                ports.notice.warn("smith did not check the selection. Update smith, or run smith compose in a terminal.")
                return false
            }
            if (check.conflict.isNotEmpty()) {
                message = say("Conflict: ${check.conflict}")
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
                                if (it.suggested) "(suggested)" else "",
                                it.description,
                            )
                        }
                val answer =
                    ports.prompt.pickOne(
                        "$TITLE: ${question.axis}",
                        "${question.choice} needs an answer on the ${question.axis} axis. ${question.about} The suggested answer is first.",
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
                message = say(if (asked == null) "${question.choice} still needs a choice on the ${question.axis} axis." else "")
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
            ports.prompt.text("$TITLE: 2 of 3, name", "The name of the new project ($stackName). The files are named after it:") {
                ProjectNameValidator.problem(it, existing)
            } ?: return false
        val planned =
            read(ports, ComposeStackAlgorithm.arguments(check.language, check.selection, kind = "app", name = name)) ?: return false
        val plan = planned.plan
        if (plan == null || plan.argv.size < 2) {
            ports.notice.warn("smith made no plan for these picks.")
            return false
        }
        val missing = ComposeStackAlgorithm.toolsToInstall(plan)
        val answer =
            ports.prompt.plan(
                "$TITLE: 3 of 3, plan",
                "Nothing is written yet. By default smith leaves the files uncommitted.",
                ComposeStackAlgorithm.describePlan(plan),
                "Add the project",
                offerCommit = true,
                others =
                    listOfNotNull(
                        if (missing.isNotEmpty()) "Install the tools first" else null,
                        if (plan.recipeArgv.size > 1) "Keep it as a recipe" else null,
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
                    ports.notice.warn("Nothing was written. ${FailureTextMapper.describe(done)}")
                    return false
                }
                ports.notice.info(done.lines.lastOrNull() ?: "The recipe is written.")
                return true
            }
        }
        val args = plan.argv.drop(1) + (if (answer.commit) listOf("--commit") else emptyList()) + "--yes"
        val done = ports.smith.run(args, show = true)
        ports.changed()
        if (!done.ok) {
            val problem = FailureTextMapper.describe(done)
            val command = TaskCommandAlgorithm.display(listOf("smith") + args)
            ports.notice.warn("Nothing was added. $problem The command was: $command")
            return false
        }
        val buttons =
            if (answer.commit) {
                emptyList()
            } else {
                listOf(
                    NoticeButtonContract("Commit now", "Make one commit that Undo can revert; you choose the files first.") {
                        commitLater("add app $name")
                    },
                )
            }
        ports.notice.info("$name is added. ${if (answer.commit) COMMITTED else UNCOMMITTED}", buttons)
        return true
    }
}
