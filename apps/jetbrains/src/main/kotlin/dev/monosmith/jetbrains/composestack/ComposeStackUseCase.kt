package dev.monosmith.jetbrains.composestack

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.addproject.ProjectNameValidator
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.installtools.InstallToolsUseCase
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithcommand.TaskCommandAlgorithm
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
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

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { compose(ide, root) }

    private fun read(
        ide: Project,
        root: Path,
        args: List<String>,
    ): ComposeReportContract? {
        val run = SmithRunClient.run(ide, root, args)
        if (!run.ok) {
            NoticeClient.warn(ide, "Nothing was added. ${FailureTextMapper.describe(run)}")
            return null
        }
        return runCatching { ComposeReportJsonMapper.toReport(run.stdout) }.getOrElse {
            NoticeClient.warn(
                ide,
                "smith printed its catalogue in a form this plugin does not read. Update the plugin, or run smith compose in a terminal.",
            )
            null
        }
    }

    private fun compose(
        ide: Project,
        root: Path,
    ) {
        val first = read(ide, root, ComposeStackAlgorithm.arguments(null, emptyList())) ?: return
        var ticked = emptyList<String>()
        var message = ""
        while (true) {
            val chosen =
                PromptClient.onUi {
                    val dialog = ComposePickDialogHandler(ide, first.layers, first.items, ticked, message, "Next: check")
                    if (dialog.showAndGet()) dialog.selected else null
                }
            if (chosen.isNullOrEmpty()) return
            val (settled, note) = ComposeStackAlgorithm.settleRivals(first.items, ticked, chosen)
            ticked = settled

            fun say(text: String) = listOf(note, text).filter { it.isNotEmpty() }.joinToString(" ")
            val ticks = ticked.mapNotNull { id -> first.items.firstOrNull { it.id == id } }
            val language =
                ticks.firstOrNull { it.kind == "language" }?.name ?: ticks.firstOrNull { it.kind == "choice" }?.language.orEmpty()
            val plugins = ticks.filter { it.kind == "plugin" }
            plugins.forEach { NoticeClient.info(ide, "${it.title} is not part of one project. Turn it on with: ${it.route}") }
            ticked = ticked.filter { id -> plugins.none { it.id == id } }
            val picks = ticks.filter { it.kind == "choice" }.map { it.name }
            if (picks.isEmpty()) {
                message = say("Tick the choice that writes the project too: a framework, a desktop shell or a cloud host.")
                continue
            }
            val checked = read(ide, root, ComposeStackAlgorithm.arguments(language, picks)) ?: return
            val check = checked.check
            if (check == null) {
                NoticeClient.warn(ide, "smith did not check the selection. Update smith, or run smith compose in a terminal.")
                return
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
                            dev.monosmith.jetbrains.prompt.ChoiceRowContract(
                                it.name,
                                it.title,
                                if (it.suggested) "(suggested)" else "",
                                it.description,
                            )
                        }
                val answer =
                    PromptClient.pickOne(
                        ide,
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
            if (note.isNotEmpty()) NoticeClient.info(ide, note)
            return finish(ide, root, check, stackName = ComposeStackAlgorithm.stackSummary(first.items, check.language, picks))
        }
    }

    private fun finish(
        ide: Project,
        root: Path,
        check: ComposeCheckContract,
        stackName: String,
    ) {
        // The name comes before the plan because the files of the plan depend on it (apps/<name>/).
        val existing = ListProjectsUseCase.list(root).getOrDefault(emptyList()).map { it.name }
        val name =
            PromptClient.text(ide, "$TITLE: 2 of 3, name", "The name of the new project ($stackName). The files are named after it:") {
                ProjectNameValidator.problem(it, existing)
            } ?: return
        while (true) {
            val planned =
                read(ide, root, ComposeStackAlgorithm.arguments(check.language, check.selection, kind = "app", name = name)) ?: return
            val plan = planned.plan
            if (plan == null || plan.argv.size < 2) return NoticeClient.warn(ide, "smith made no plan for these picks.")
            val missing = ComposeStackAlgorithm.toolsToInstall(plan)
            val answer =
                PromptClient.plan(
                    ide,
                    "$TITLE: 3 of 3, plan",
                    "Nothing is written yet. By default smith leaves the files uncommitted.",
                    ComposeStackAlgorithm.describePlan(plan),
                    "Add the project",
                    offerCommit = true,
                    others =
                        listOfNotNull(
                            if (missing.isNotEmpty()) "Install the tools first" else null,
                            if (plan.recipeArgv.size >
                                1
                            ) {
                                "Keep it as a recipe"
                            } else {
                                null
                            },
                        ),
                ) ?: return
            val others = listOfNotNull(if (missing.isNotEmpty()) "tools" else null, if (plan.recipeArgv.size > 1) "recipe" else null)
            when (others.getOrNull(answer.other)) {
                "tools" -> {
                    return InstallToolsUseCase.start(ide, root)
                }

                "recipe" -> {
                    val done = SmithRunClient.run(ide, root, plan.recipeArgv.drop(1), show = true)
                    return if (done.ok) {
                        NoticeClient.info(
                            ide,
                            done.lines.lastOrNull() ?: "The recipe is written.",
                        )
                    } else {
                        NoticeClient.warn(ide, "Nothing was written. ${FailureTextMapper.describe(done)}")
                    }
                }
            }
            val args = plan.argv.drop(1) + (if (answer.commit) listOf("--commit") else emptyList()) + "--yes"
            val done = SmithRunClient.run(ide, root, args, show = true)
            WorkspaceChangedStore.changed(ide)
            if (done.ok) {
                val commit = if (answer.commit) COMMITTED else UNCOMMITTED
                val buttons =
                    if (answer.commit) {
                        emptyList()
                    } else {
                        listOf(
                            NoticeButtonContract(
                                "Commit now",
                                "Make one commit that Undo can revert; you choose the files first.",
                            ) {
                                CommitChangesUseCase.start(ide, root, "add app $name")
                            },
                        )
                    }
                NoticeClient.info(ide, "$name is added. $commit", buttons)
            } else {
                NoticeClient.warn(
                    ide,
                    "Nothing was added. ${FailureTextMapper.describe(
                        done,
                    )} The command was: ${TaskCommandAlgorithm.display(listOf("smith") + args)}",
                )
            }
            return
        }
    }
}
