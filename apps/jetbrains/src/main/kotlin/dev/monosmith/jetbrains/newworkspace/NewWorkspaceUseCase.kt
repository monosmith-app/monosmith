package dev.monosmith.jetbrains.newworkspace

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.addproject.ProjectNameValidator
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Creates a MonoSmith workspace in the folder the IDE has open. First question: "How do you want to set
 * it up?" with the defaults (the line lists what they choose) or "Customise" (every question of
 * `smith new`, read from `smith new --questions --json`, so the lists are smith's own). A summary and the
 * files of `smith new --dry-run` come before anything is written.
 */
object NewWorkspaceUseCase {
    private const val TITLE = "MonoSmith: new workspace"

    /** Starts the flow in a background task; [folder] is the open folder. */
    fun start(
        ide: Project,
        folder: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, folder), folder.fileName?.toString().orEmpty()) }

    /**
     * @param folderName the name of the open folder, offered as the workspace name
     * @return true when the workspace was created
     */
    fun run(
        ports: FlowPortsContract,
        folderName: String,
    ): Boolean {
        val listed = ports.smith.run(listOf("new", "--questions", "--json"))
        val questions =
            runCatching { NewQuestionsMapper.toQuestions(listed.stdout) }.getOrElse {
                ports.notice.warn(
                    "Nothing was created. smith did not list its questions: ${FailureTextMapper.describe(listed)} Update smith.",
                )
                return false
            }
        val defaults = NewWorkspaceArgumentsAlgorithm.defaults(folderName, questions)
        val way =
            ports.prompt
                .pickOne(
                    "$TITLE: how do you want to set it up?",
                    "Nothing is written before you see the files.",
                    listOf(
                        ChoiceRowContract(
                            "defaults",
                            "Start with the defaults",
                            "(default)",
                            NewWorkspaceArgumentsAlgorithm.describe(defaults, folderName, questions).drop(1).joinToString("; "),
                        ),
                        ChoiceRowContract("customise", "Customise", "", "Answer each question of smith new."),
                    ),
                )?.name ?: return false
        val name =
            ports.prompt.text("$TITLE: name", "The name of the workspace (the name of this folder is $folderName):", folderName) {
                ProjectNameValidator.problem(it, emptyList())
            } ?: return false
        val choices =
            if (way ==
                "customise"
            ) {
                customise(ports, questions, defaults.copy(name = name)) ?: return false
            } else {
                defaults.copy(name = name)
            }
        val preview = ports.smith.run(NewWorkspaceArgumentsAlgorithm.arguments(choices, preview = true))
        if (!preview.ok) {
            ports.notice.warn("Nothing was created. ${FailureTextMapper.describe(preview)}")
            return false
        }
        val summary = NewWorkspaceArgumentsAlgorithm.describe(choices, folderName, questions).joinToString("\n")
        ports.prompt.plan(
            "$TITLE: create it?",
            "Nothing is written yet. These are your choices and the files smith would write.",
            "$summary\n\n${preview.stdout.trim()}",
            "Create the workspace",
            offerCommit = false,
        )
            ?: return false
        val done = ports.smith.run(NewWorkspaceArgumentsAlgorithm.arguments(choices, preview = false), show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn("Nothing was created. ${FailureTextMapper.describe(done)}")
            return false
        }
        ports.notice.info("${choices.name} is created.")
        return true
    }

    private fun customise(
        ports: FlowPortsContract,
        questions: List<NewQuestionContract>,
        start: NewWorkspaceChoicesContract,
    ): NewWorkspaceChoicesContract? {
        val answers = mutableMapOf<String, String>()
        for (question in NewQuestionsMapper.main(questions)) {
            if (!NewQuestionsMapper.applies(question, answers)) continue
            val rows = NewQuestionsMapper.rows(question)
            val title = "$TITLE: ${question.key}"
            answers[question.key] =
                if (question.many) {
                    (ports.prompt.pickMany(title, "${question.prompt} The default is ticked.", rows) ?: return null)
                        .ifEmpty {
                            listOf(question.default)
                        }.joinToString(",")
                } else {
                    ports.prompt.pickOne(title, "${question.prompt} The first answer is the default.", rows)?.name ?: return null
                }
        }
        val commit =
            ports.prompt
                .pickOne(
                    "$TITLE: first commit",
                    "Make the first commit of smith's own files? The first answer is the default.",
                    NewWorkspaceArgumentsAlgorithm.yesNoRows(
                        "smith makes the first commit of its own files.",
                        "The files stay uncommitted.",
                        true,
                    ),
                )?.name ?: return null
        return start.copy(
            ci = answers["ci"]?.split(",").orEmpty().ifEmpty { start.ci },
            language = answers["language"] ?: start.language,
            nodeStack = answers["node-stack"] ?: start.nodeStack,
            architecture = answers["architecture"] ?: start.architecture,
            agent = (answers["agent"] ?: "yes") == "yes",
            devcontainer = (answers["devcontainer"] ?: "yes") == "yes",
            commit = commit == "yes",
        )
    }
}
