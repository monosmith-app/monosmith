package dev.monosmith.jetbrains.newworkspace

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.addproject.ProjectNameValidator
import dev.monosmith.jetbrains.notice.NoticeClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.prompt.PromptClient
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.smithrun.SmithRunClient
import dev.monosmith.jetbrains.workspace.WorkspaceChangedStore
import java.nio.file.Path

/**
 * Creates a MonoSmith workspace in the folder the IDE has open. First question: "How do you want to set
 * it up?" with the defaults (the line lists what they choose) or "Customise" (every question of
 * `smith new`). A summary and the files of `smith new --dry-run` come before anything is written.
 */
object NewWorkspaceUseCase {
    private const val TITLE = "MonoSmith: new workspace"

    /** Starts the flow in a background task; [folder] is the open folder. */
    fun start(
        ide: Project,
        folder: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) {
        val folderName = folder.fileName?.toString().orEmpty()
        val defaults = NewWorkspaceChoicesContract(folderName)
        val way =
            PromptClient
                .pickOne(
                    ide,
                    "$TITLE: how do you want to set it up?",
                    "Nothing is written before you see the files.",
                    listOf(
                        ChoiceRowContract(
                            "defaults",
                            "Start with the defaults",
                            "(default)",
                            NewWorkspaceArgumentsAlgorithm.describe(defaults, folderName).drop(1).joinToString("; "),
                        ),
                        ChoiceRowContract(
                            "customise",
                            "Customise",
                            "",
                            "Answer each question of smith new: CI, language, layout, agent files, dev container, first commit.",
                        ),
                    ),
                )?.name ?: return@start
        val name =
            PromptClient.text(ide, "$TITLE: name", "The name of the workspace (the name of this folder is $folderName):", folderName) {
                ProjectNameValidator.problem(it, emptyList())
            } ?: return@start
        val choices =
            if (way ==
                "customise"
            ) {
                customise(ide, NewWorkspaceChoicesContract(name)) ?: return@start
            } else {
                NewWorkspaceChoicesContract(name)
            }
        val preview = SmithRunClient.run(ide, folder, NewWorkspaceArgumentsAlgorithm.arguments(choices, preview = true))
        if (!preview.ok) return@start NoticeClient.warn(ide, "Nothing was created. ${FailureTextMapper.describe(preview)}")
        val summary = NewWorkspaceArgumentsAlgorithm.describe(choices, folderName).joinToString("\n")
        PromptClient.plan(
            ide,
            "$TITLE: create it?",
            "Nothing is written yet. These are your choices and the files smith would write.",
            "$summary\n\n${preview.stdout.trim()}",
            "Create the workspace",
            offerCommit = false,
        )
            ?: return@start
        val done = SmithRunClient.run(ide, folder, NewWorkspaceArgumentsAlgorithm.arguments(choices, preview = false), show = true)
        WorkspaceChangedStore.changed(ide)
        if (done.ok) {
            NoticeClient.info(
                ide,
                "${choices.name} is created.",
            )
        } else {
            NoticeClient.warn(ide, "Nothing was created. ${FailureTextMapper.describe(done)}")
        }
    }

    private fun customise(
        ide: Project,
        start: NewWorkspaceChoicesContract,
    ): NewWorkspaceChoicesContract? {
        val ci =
            PromptClient.pickMany(
                ide,
                "$TITLE: CI",
                "CI = where your tests run on every push. The default is ticked.",
                NewWorkspaceArgumentsAlgorithm.ciRows,
            )
                ?: return null
        val language =
            PromptClient
                .pickOne(
                    ide,
                    "$TITLE: main language",
                    "The first answer is the default.",
                    NewWorkspaceArgumentsAlgorithm.languageRows,
                )?.name
                ?: return null
        val nodeStack =
            if (language == "node") {
                PromptClient
                    .pickOne(
                        ide,
                        "$TITLE: Node tests and lint",
                        "The first answer is the default.",
                        NewWorkspaceArgumentsAlgorithm.nodeStackRows,
                    )?.name
                    ?: return null
            } else {
                start.nodeStack
            }
        val architecture =
            PromptClient
                .pickOne(
                    ide,
                    "$TITLE: code layout",
                    "smith checks the layout you pick. The first answer is the default.",
                    NewWorkspaceArgumentsAlgorithm.architectureRows,
                )?.name
                ?: return null
        val agent =
            yesNo(ide, "agent files", "Write the MonoSmith agent files for Claude Code.", "Write no agent files.", true) ?: return null
        val container = yesNo(ide, "dev container", "Write a dev container.", "Write no dev container.", true) ?: return null
        val commit =
            yesNo(ide, "first commit", "smith makes the first commit of its own files.", "The files stay uncommitted.", true) ?: return null
        return start.copy(
            ci =
                ci.ifEmpty {
                    start.ci
                },
            language = language,
            nodeStack = nodeStack,
            architecture = architecture,
            agent = agent,
            devcontainer = container,
            commit = commit,
        )
    }

    private fun yesNo(
        ide: Project,
        what: String,
        yes: String,
        no: String,
        defaultYes: Boolean,
    ): Boolean? =
        PromptClient
            .pickOne(
                ide,
                "$TITLE: $what",
                "The first answer is the default.",
                NewWorkspaceArgumentsAlgorithm.yesNoRows(yes, no, defaultYes),
            )?.name
            ?.let {
                it ==
                    "yes"
            }
}
