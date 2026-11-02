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
import java.nio.file.Path

/**
 * Converts an Nx or Turborepo workspace (`smith convert`): the level first, at the full level the four
 * choices behind it, then the plan of `--dry-run` that lists the files it deletes, and only then `--yes`.
 * The questions and their answers come from `smith convert --questions --json`; the plugin holds no list of
 * them. smith commits only when the box is ticked.
 */
object ConvertFolderUseCase {
    private const val TITLE = "MonoSmith: convert this workspace"

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
            ports.notice.warn("Nothing was changed. ${FailureTextMapper.describe(listed)} Update smith.")
            return false
        }
        val questions =
            runCatching { NewQuestionsMapper.toQuestions(listed.stdout) }.getOrElse {
                ports.notice.warn("smith printed the questions of convert in a form this plugin does not read. Update the plugin.")
                return false
            }
        val answers = mutableMapOf<String, String>()
        for (question in questions) {
            if (question.options.isEmpty() || !NewQuestionsMapper.applies(question, answers)) continue
            val header = if (question.key == "level") "The first answer changes the least." else "The first answer is the default."
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
                header = "Nothing was changed yet. The plan lists the files smith deletes and the changes it makes.",
                okText = "Convert",
                offerCommit = true,
                apply = { ConvertArgumentsAlgorithm.arguments(chosen, apply = true) },
                done = "The workspace is converted",
                operation = "convert",
            ),
            commitLater,
        )
    }
}
