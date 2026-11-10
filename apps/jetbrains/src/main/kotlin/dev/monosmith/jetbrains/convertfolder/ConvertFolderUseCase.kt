package dev.monosmith.jetbrains.convertfolder

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.commitchanges.CommitChangesUseCase
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.newworkspace.NewQuestionsMapper
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.ADOPT_REVIEW_CONVERT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CONVERTFOLDER_MONOSMITH_CONVERT_THIS_WORKSPACE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CONVERTFOLDER_NOTHING_WAS_CHANGED_UPDATE_SMITH
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CONVERTFOLDER_NOTHING_WAS_CHANGED_YET_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CONVERTFOLDER_SMITH_PRINTED_THE_QUESTIONS_OF
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CONVERTFOLDER_THE_FIRST_ANSWER_CHANGES_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_CONVERTFOLDER_THE_WORKSPACE_IS_CONVERTED
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Converts an Nx or Turborepo workspace (`smith convert`): the level first, at the full level the four
 * choices behind it, then the plan of `--dry-run` that lists the files it deletes, and only then `--yes`.
 * The questions and their answers come from `smith convert --questions --json`; the plugin holds no list of
 * them. smith commits only when the box is ticked.
 */
object ConvertFolderUseCase {
    private val TITLE = say(JETBRAINS_CONVERTFOLDER_MONOSMITH_CONVERT_THIS_WORKSPACE)

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) { CommitChangesUseCase.start(ide, root, it) } }

    /**
     * @return true when smith converted the workspace
     */
    fun run(
        ports: FlowPortsContract,
        commitLater: (String) -> Unit,
    ): Boolean {
        val listed = ports.smith.run(listOf("convert", "--questions", "--json"))
        if (!listed.ok) {
            ports.notice.warn(say(JETBRAINS_CONVERTFOLDER_NOTHING_WAS_CHANGED_UPDATE_SMITH, FailureTextMapper.describe(listed).toString()))
            return false
        }
        val questions =
            runCatching { NewQuestionsMapper.toQuestions(listed.stdout) }.getOrElse {
                ports.notice.warn(say(JETBRAINS_CONVERTFOLDER_SMITH_PRINTED_THE_QUESTIONS_OF))
                return false
            }
        val answers = mutableMapOf<String, String>()
        for (question in questions) {
            if (question.options.isEmpty() || !NewQuestionsMapper.applies(question, answers)) continue
            val header =
                if (question.key ==
                    "level"
                ) {
                    say(JETBRAINS_CONVERTFOLDER_THE_FIRST_ANSWER_CHANGES_THE)
                } else {
                    say(JETBRAINS_ADDPROJECT_THE_FIRST_ANSWER_IS_THE)
                }
            answers[question.key] =
                ports.prompt.pickOne("$TITLE: ${question.key}", "${question.prompt}\n$header", NewQuestionsMapper.rows(question))?.name
                    ?: return false
        }
        val defaults = ConvertChoicesContract()
        val chosen =
            ConvertChoicesContract(
                level = answers["level"] ?: defaults.level,
                lint = answers["lint"] ?: defaults.lint,
                test = answers["test"] ?: defaults.test,
                build = answers["build"] ?: defaults.build,
                ci = answers["ci"] ?: defaults.ci,
            )
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE?",
                planArgs = ConvertArgumentsAlgorithm.arguments(chosen, apply = false),
                header = say(JETBRAINS_CONVERTFOLDER_NOTHING_WAS_CHANGED_YET_THE),
                okText = say(ADOPT_REVIEW_CONVERT),
                offerCommit = true,
                apply = { ConvertArgumentsAlgorithm.arguments(chosen, apply = true) },
                done = say(JETBRAINS_CONVERTFOLDER_THE_WORKSPACE_IS_CONVERTED),
                operation = "convert",
            ),
            commitLater,
        )
    }
}
