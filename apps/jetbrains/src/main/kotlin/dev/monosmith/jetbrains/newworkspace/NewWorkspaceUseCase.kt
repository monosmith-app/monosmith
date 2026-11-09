package dev.monosmith.jetbrains.newworkspace

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.addproject.ProjectNameValidator
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_SETUP_CUSTOMISE
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_ADD_PROJECT_SETUP_START_WITH_THE_DEFAULTS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_HOW_DO_YOU_WANT_TO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_ANSWER_EACH_QUESTION_OF_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_CREATE_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_CREATE_THE_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_FIRST_COMMIT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_IS_CREATED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_MAKE_THE_FIRST_COMMIT_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_MONOSMITH_NEW_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_NOTHING_IS_WRITTEN_BEFORE_YOU
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_NOTHING_IS_WRITTEN_YET_THESE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_NOTHING_WAS_CREATED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_NOTHING_WAS_CREATED_SMITH_DID
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_SMITH_MAKES_THE_FIRST_COMMIT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_THE_DEFAULT_IS_TICKED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_THE_FILES_STAY_UNCOMMITTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_THE_FIRST_ANSWER_IS_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_NEWWORKSPACE_THE_NAME_OF_THE_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Creates a MonoSmith workspace in the folder the IDE has open. First question: "How do you want to set
 * it up?" with the defaults (the line lists what they choose) or "Customise" (every question of
 * `smith new`, read from `smith new --questions --json`, so the lists are smith's own). A summary and the
 * files of `smith new --dry-run` come before anything is written.
 */
object NewWorkspaceUseCase {
    private val TITLE = say(JETBRAINS_NEWWORKSPACE_MONOSMITH_NEW_WORKSPACE)

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
                    say(JETBRAINS_NEWWORKSPACE_NOTHING_WAS_CREATED_SMITH_DID, FailureTextMapper.describe(listed).toString()),
                )
                return false
            }
        val defaults = NewWorkspaceArgumentsAlgorithm.defaults(folderName, questions)
        val way =
            ports.prompt
                .pickOne(
                    say(JETBRAINS_ADDPROJECT_HOW_DO_YOU_WANT_TO, TITLE.toString()),
                    say(JETBRAINS_NEWWORKSPACE_NOTHING_IS_WRITTEN_BEFORE_YOU),
                    listOf(
                        ChoiceRowContract(
                            "defaults",
                            say(GUIDE_ADD_PROJECT_SETUP_START_WITH_THE_DEFAULTS),
                            say(JETBRAINS_ADDCHOICES_DEFAULT),
                            NewWorkspaceArgumentsAlgorithm.describe(defaults, folderName, questions).drop(1).joinToString("; "),
                        ),
                        ChoiceRowContract(
                            "customise",
                            say(GUIDE_ADD_PROJECT_SETUP_CUSTOMISE),
                            "",
                            say(JETBRAINS_NEWWORKSPACE_ANSWER_EACH_QUESTION_OF_SMITH),
                        ),
                    ),
                )?.name ?: return false
        val name =
            ports.prompt.text("$TITLE: name", say(JETBRAINS_NEWWORKSPACE_THE_NAME_OF_THE_WORKSPACE, folderName.toString()), folderName) {
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
            ports.notice.warn(say(JETBRAINS_NEWWORKSPACE_NOTHING_WAS_CREATED, FailureTextMapper.describe(preview).toString()))
            return false
        }
        val summary = NewWorkspaceArgumentsAlgorithm.describe(choices, folderName, questions).joinToString("\n")
        ports.prompt.plan(
            say(JETBRAINS_NEWWORKSPACE_CREATE_IT, TITLE.toString()),
            say(JETBRAINS_NEWWORKSPACE_NOTHING_IS_WRITTEN_YET_THESE),
            "$summary\n\n${preview.stdout.trim()}",
            say(JETBRAINS_NEWWORKSPACE_CREATE_THE_WORKSPACE),
            offerCommit = false,
        )
            ?: return false
        val done = ports.smith.run(NewWorkspaceArgumentsAlgorithm.arguments(choices, preview = false), show = true)
        ports.changed()
        if (!done.ok) {
            ports.notice.warn(say(JETBRAINS_NEWWORKSPACE_NOTHING_WAS_CREATED, FailureTextMapper.describe(done).toString()))
            return false
        }
        ports.notice.info(say(JETBRAINS_NEWWORKSPACE_IS_CREATED, choices.name.toString()))
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
                    (
                        ports.prompt.pickMany(title, say(JETBRAINS_NEWWORKSPACE_THE_DEFAULT_IS_TICKED, question.prompt.toString()), rows)
                            ?: return null
                    ).ifEmpty {
                        listOf(question.default)
                    }.joinToString(",")
                } else {
                    ports.prompt
                        .pickOne(
                            title,
                            say(JETBRAINS_NEWWORKSPACE_THE_FIRST_ANSWER_IS_THE, question.prompt.toString()),
                            rows,
                        )?.name ?: return null
                }
        }
        val commit =
            ports.prompt
                .pickOne(
                    say(JETBRAINS_NEWWORKSPACE_FIRST_COMMIT, TITLE.toString()),
                    say(JETBRAINS_NEWWORKSPACE_MAKE_THE_FIRST_COMMIT_OF),
                    NewWorkspaceArgumentsAlgorithm.yesNoRows(
                        say(JETBRAINS_NEWWORKSPACE_SMITH_MAKES_THE_FIRST_COMMIT),
                        say(JETBRAINS_NEWWORKSPACE_THE_FILES_STAY_UNCOMMITTED),
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
