package dev.monosmith.jetbrains.addproject

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.addchoices.ChoicesJsonMapper
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.installdependency.InstallDependencyUseCase
import dev.monosmith.jetbrains.installtools.InstallToolsUseCase
import dev.monosmith.jetbrains.notice.NoticeButtonContract
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.OutputConsoleStore
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.smithrun.SmithRunContract
import dev.monosmith.jetbrains.undochange.UndoChangeUseCase
import dev.monosmith.jetbrains.workspace.ListProjectsUseCase
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
import java.nio.file.Path

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
        val options = readOptions(ide, root) ?: return@start
        val siblings = ListProjectsUseCase.list(root).getOrDefault(emptyList()).map { AddQuestionsUseCase.Sibling(it.name, it.language) }
        val answer = AddQuestionsUseCase.ask(ide, options, siblings) ?: return@start
        if (answer is AddAnswerContract.Compose) return@start onCompose()
        val (request, words) = answer as AddAnswerContract.Request
        val plan = SmithRunClient.run(ide, root, AddArgumentsAlgorithm.arguments(request, apply = false))
        if (!plan.ok) return@start NoticeClient.warn(ide, "Nothing was added. ${FailureTextMapper.describe(plan)}")
        val follows = AddArgumentsAlgorithm.followsSiblings(plan.stdout)
        val header = "$words Nothing was changed yet. By default smith leaves the files uncommitted: review them with the Commit window."
        val answered =
            PromptClient.plan(
                ide,
                TITLE,
                header,
                plan.stdout.trim(),
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
            ) ?: return@start
        val fresh = request.fresh || answered.checks.firstOrNull() == true
        val done =
            SmithRunClient.run(
                ide,
                root,
                AddArgumentsAlgorithm.arguments(request, apply = true, fresh = fresh, commit = answered.commit),
                show = true,
            )
        WorkspaceChangedStore.changed(ide)
        tell(ide, root, request, answered.commit, done)
    }

    private fun readOptions(
        ide: Project,
        root: Path,
    ): AddOptionsContract? {
        val kinds = SmithRunClient.run(ide, root, listOf("kinds", "--json"))
        val choices = SmithRunClient.run(ide, root, listOf("choices", "--json"))
        val catalogue = SmithRunClient.run(ide, root, listOf("compose", "--json"))
        val failed = listOf(kinds, choices, catalogue).firstOrNull { !it.ok }
        if (failed != null) {
            NoticeClient.warn(ide, "Nothing was added. ${FailureTextMapper.describe(failed)}")
            return null
        }
        return runCatching {
            val kindList = OutcomesJsonMapper.toKinds(kinds.stdout)
            val (outcomes, languages, layers) = OutcomesJsonMapper.toCatalogue(catalogue.stdout)
            AddOptionsContract(
                kindList,
                ChoicesJsonMapper.toChoices(choices.stdout),
                OutcomesJsonMapper.fitOutcomes(outcomes, kindList),
                languages,
                layers,
            )
        }.getOrElse {
            NoticeClient.warn(
                ide,
                "Nothing was added. smith printed its lists in a form this plugin does not read. Update the plugin, or run smith add in a terminal.",
            )
            null
        }
    }

    private fun tell(
        ide: Project,
        root: Path,
        request: AddRequestContract,
        committed: Boolean,
        done: SmithRunContract,
    ) {
        val operation = "add ${request.role} ${request.name}"
        if (done.ok) {
            val lib = request.role == "lib"
            val buttons =
                listOfNotNull(
                    if (committed) {
                        NoticeButtonContract(
                            "Undo",
                            "Revert the commit of this add.",
                        ) { UndoChangeUseCase.start(ide, root) }
                    } else {
                        null
                    },
                    if (committed) {
                        null
                    } else {
                        NoticeButtonContract("Commit now", "Make one commit that Undo can revert; you choose the files first.") {
                            CommitChangesUseCase.start(ide, root, operation)
                        }
                    },
                    if (lib) {
                        NoticeButtonContract(
                            "Use in an app…",
                            "Add this library to an app.",
                        ) { InstallDependencyUseCase.start(ide, root, only = "workspace") }
                    } else {
                        null
                    },
                )
            val commit = if (committed) COMMITTED else UNCOMMITTED
            NoticeClient.info(ide, "${request.name} is added and the CI pipelines are regenerated. $commit", buttons)
            return
        }
        val everything = "${done.stdout}\n${done.stderr}"
        val failure = AddFailureMapper.doctorFailure(everything)
        if (failure != null) return tellFailing(ide, root, operation, failure)
        val tool = AddFailureMapper.missingTool(everything)
        if (tool != null) return tellUnchecked(ide, root, request.name, tool)
        NoticeClient.warn(ide, "Nothing was added. ${FailureTextMapper.describe(done)}")
    }

    private fun tellFailing(
        ide: Project,
        root: Path,
        operation: String,
        failure: DoctorFailureContract,
    ) = NoticeClient.warn(
        ide,
        AddFailureMapper.describe(failure),
        listOf(
            NoticeButtonContract(
                "Install tools",
                "Install the missing tools with smith tools install.",
            ) { InstallToolsUseCase.start(ide, root) },
            NoticeButtonContract(
                "Open Health",
                "Open the Health tab of the MonoSmith tool window.",
            ) { OutputConsoleStore.reveal(ide, "Health") },
            NoticeButtonContract("Commit anyway", "Commit the project although it fails its checks; you choose the files first.") {
                CommitChangesUseCase.start(ide, root, operation)
            },
            NoticeButtonContract("Keep uncommitted", "Leave the files in your work tree.") {},
        ),
    )

    private fun tellUnchecked(
        ide: Project,
        root: Path,
        name: String,
        tool: MissingToolContract,
    ) {
        val where = tool.dir?.let { " in $it" }.orEmpty()
        val files = if (tool.files > 0) " (${tool.files} files)" else ""
        val check = tool.check
        val buttons =
            listOfNotNull(
                tool.url?.let { url ->
                    NoticeButtonContract("Install page", "Open the page that tells how to get ${tool.title}.") { BrowserUtil.browse(url) }
                },
                NoticeButtonContract(
                    "Install tools",
                    "Install the missing tools with smith tools install.",
                ) { InstallToolsUseCase.start(ide, root) },
                if (check == null) {
                    null
                } else {
                    NoticeButtonContract("Check again", "Run the first check of the new project again.") {
                        BackgroundFlowClient.start(ide, "MonoSmith: check $name") { SmithRunClient.run(ide, root, check, show = true) }
                    }
                },
                NoticeButtonContract("Skip the check", "Continue without the check. The Health tab keeps reminding you.") {},
            )
        NoticeClient.warn(
            ide,
            "$name was created$where$files, and could not be checked yet: ${tool.title} is not installed. Nothing is lost: your project stays on disk.",
            buttons,
        )
    }
}
