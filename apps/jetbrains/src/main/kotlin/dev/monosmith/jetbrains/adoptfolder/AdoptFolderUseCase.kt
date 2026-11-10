package dev.monosmith.jetbrains.adoptfolder

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.newworkspace.NewQuestionsMapper
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.ADOPT_AGENT_DECLINE
import dev.monosmith.jetbrains.words.SmithWordsConfig.GUIDE_NEW_WORKSPACE_SETUP_QUESTION_AGENT_NAME
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_ADD_A_GITATTRIBUTES_WITH_TEXT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_ADD_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_ADD_THEM
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_ADD_THE_MONOSMITH_AGENT_FILES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_BRING_IT_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_IS_AT_THE_ROOT_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_KEEP_MINE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_LEAVE_LINE_ENDINGS_ALONE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_LINE_ENDINGS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_MONOSMITH_BRING_THIS_FOLDER_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_NEVER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_NEXT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_NOTHING_WAS_ADOPTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_NOTHING_WAS_ADOPTED_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_NOTHING_WAS_CHANGED_YET_THIS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_RECORDS_AGENT_ENABLED_FALSE_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_REPLACE_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_SMITH_WRITES_ITS_OWN_YOURS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_SMITH_WRITES_NO_PIPELINE_FOR
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_THE_FIRST_ANSWER_IS_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_THE_FOLDER_FROM_THE_REPOSITORY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_THE_FOLDER_IS_A_MONOSMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_THE_NAME_IN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_TICK_NONE_TO_WRITE_NO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_WHERE_DOES_GO
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_WRITES_CLAUDE_AGENTS_SMITH_MD
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_YOUR_PIPELINES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADOPTFOLDER_YOU_CAN_ADD_THEM_LATER
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Brings a repository that has no smith.toml into MonoSmith (`smith adopt --init`). smith reports what it
 * found (`--dry-run`); the person then answers the four questions a terminal asks (line endings, agent
 * files, CI providers, a pipeline of their own) and, for a project at the repository root, where to move
 * it; the plan with diffs comes next, and only then smith writes. Every answer is a flag, so nothing is
 * asked twice; nothing is committed unless the box is ticked.
 */
object AdoptFolderUseCase {
    private val TITLE = say(JETBRAINS_ADOPTFOLDER_MONOSMITH_BRING_THIS_FOLDER_IN)

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) } }

    /**
     * @return true when smith adopted the folder
     */
    fun run(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val found = ports.smith.run(listOf("adopt", "--dry-run", "--json"))
        val report =
            runCatching { AdoptReportMapper.toReport(found.stdout) }.getOrElse {
                ports.notice.warn(say(JETBRAINS_ADOPTFOLDER_NOTHING_WAS_ADOPTED, FailureTextMapper.describe(found).toString()))
                return false
            }
        if (report.blockers.isNotEmpty()) {
            ports.notice.warn(say(JETBRAINS_ADOPTFOLDER_NOTHING_WAS_ADOPTED_2, report.blockers.first().toString()))
            return false
        }
        var move: String? = null
        report.rootProject?.let { project ->
            val choice =
                ports.prompt
                    .pickOne(
                        say(JETBRAINS_ADOPTFOLDER_WHERE_DOES_GO, TITLE.toString(), project.name.toString()),
                        say(JETBRAINS_ADOPTFOLDER_IS_AT_THE_ROOT_SMITH, project.manifest.toString()),
                        AdoptReportMapper.moveRows(project),
                        say(JETBRAINS_ADOPTFOLDER_NEXT),
                    )?.name
                    ?: return false
            move = askMove(ports, project, choice) ?: return false
        }
        val gitattributes =
            pick(
                ports,
                say(JETBRAINS_ADOPTFOLDER_LINE_ENDINGS),
                say(JETBRAINS_ADOPTFOLDER_ADD_A_GITATTRIBUTES_WITH_TEXT),
                yesRows("add", "skip", say(JETBRAINS_ADOPTFOLDER_ADD_IT), say(JETBRAINS_ADOPTFOLDER_LEAVE_LINE_ENDINGS_ALONE)),
            )
                ?: return false
        val agent =
            pick(
                ports,
                say(GUIDE_NEW_WORKSPACE_SETUP_QUESTION_AGENT_NAME),
                say(JETBRAINS_ADOPTFOLDER_ADD_THE_MONOSMITH_AGENT_FILES),
                listOf(
                    ChoiceRowContract(
                        "add",
                        say(JETBRAINS_ADOPTFOLDER_ADD_THEM),
                        say(JETBRAINS_ADDCHOICES_DEFAULT),
                        say(JETBRAINS_ADOPTFOLDER_WRITES_CLAUDE_AGENTS_SMITH_MD),
                    ),
                    ChoiceRowContract("skip", say(ADOPT_AGENT_DECLINE), "", say(JETBRAINS_ADOPTFOLDER_YOU_CAN_ADD_THEM_LATER)),
                    ChoiceRowContract(
                        "never",
                        say(JETBRAINS_ADOPTFOLDER_NEVER),
                        "",
                        say(JETBRAINS_ADOPTFOLDER_RECORDS_AGENT_ENABLED_FALSE_IN),
                    ),
                ),
            ) ?: return false
        val providers = askProviders(ports) ?: return false
        val existing =
            if (report.providers.isEmpty() || providers.isEmpty()) {
                null
            } else {
                pick(
                    ports,
                    say(JETBRAINS_ADOPTFOLDER_YOUR_PIPELINES),
                    "The repository already has a pipeline for ${report.providers.joinToString(
                        ", ",
                    )}. Replace it with smith's, or keep it?",
                    listOf(
                        ChoiceRowContract(
                            "keep",
                            say(JETBRAINS_ADOPTFOLDER_KEEP_MINE),
                            say(JETBRAINS_ADDCHOICES_DEFAULT),
                            say(JETBRAINS_ADOPTFOLDER_SMITH_WRITES_NO_PIPELINE_FOR),
                        ),
                        ChoiceRowContract(
                            "replace",
                            say(JETBRAINS_ADOPTFOLDER_REPLACE_IT),
                            "",
                            say(JETBRAINS_ADOPTFOLDER_SMITH_WRITES_ITS_OWN_YOURS),
                        ),
                    ),
                ) ?: return false
            }
        val plan = AdoptReportMapper.initArguments(move, gitattributes, agent, providers, existing, apply = false)
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = plan,
                header = say(JETBRAINS_ADOPTFOLDER_NOTHING_WAS_CHANGED_YET_THIS),
                okText = say(JETBRAINS_ADOPTFOLDER_BRING_IT_IN),
                offerCommit = true,
                apply = { AdoptReportMapper.initArguments(move, gitattributes, agent, providers, existing, apply = true) },
                done = say(JETBRAINS_ADOPTFOLDER_THE_FOLDER_IS_A_MONOSMITH),
                operation = "adopt",
            ),
            commitLater,
        )
    }

    private fun yesRows(
        yes: String,
        no: String,
        yesText: String,
        noText: String,
    ) = listOf(ChoiceRowContract(yes, yesText, say(JETBRAINS_ADDCHOICES_DEFAULT), ""), ChoiceRowContract(no, noText, "", ""))

    private fun pick(
        ports: FlowPortsContract,
        what: String,
        header: String,
        rows: List<ChoiceRowContract>,
    ): String? =
        ports.prompt.pickOne("$TITLE: $what", say(JETBRAINS_ADOPTFOLDER_THE_FIRST_ANSWER_IS_THE, header.toString()), rows, "Next")?.name

    private fun askMove(
        ports: FlowPortsContract,
        project: RootProjectContract,
        choice: String,
    ): String? {
        if (choice == "none") return AdoptReportMapper.LEAVE
        val whole = choice == "other"
        val typed =
            ports.prompt.text(
                "$TITLE: ${if (whole) "folder" else "name"}",
                if (whole) {
                    say(
                        JETBRAINS_ADOPTFOLDER_THE_FOLDER_FROM_THE_REPOSITORY,
                        project.name.toString(),
                    )
                } else {
                    say(JETBRAINS_ADOPTFOLDER_THE_NAME_IN, choice.toString())
                },
                if (whole) "" else project.name,
            ) { if (whole) AdoptReportMapper.folderOrProblem(it).second else AdoptReportMapper.nameProblem(it) } ?: return null
        return if (whole) AdoptReportMapper.folderOrProblem(typed).first else "$choice/$typed"
    }

    private fun askProviders(ports: FlowPortsContract): List<String>? {
        val listed = ports.smith.run(listOf("new", "--questions", "--json"))
        val question = runCatching { NewQuestionsMapper.toQuestions(listed.stdout).first { it.key == "ci" } }.getOrNull()
        val rows = question?.let { NewQuestionsMapper.rows(it) } ?: return emptyList()
        return ports.prompt.pickMany("$TITLE: CI", say(JETBRAINS_ADOPTFOLDER_TICK_NONE_TO_WRITE_NO, question.prompt.toString()), rows)
    }
}
