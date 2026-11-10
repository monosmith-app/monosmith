package dev.monosmith.jetbrains.pipelinetemplate

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_MONOSMITH_ADD_A_PIPELINE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_NAME_THE_APP
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_NO_PIPELINE_WAS_ADDED
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_PICK_A_PIPELINE_IT_IS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_SMITH_LISTED_NO_PIPELINE_UPDATE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_THE_APP_THIS_PIPELINE_BUILDS
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_THE_FILE_IS_YOURS_FROM
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_THE_PIPELINE_IS_WRITTEN
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_PIPELINETEMPLATE_WRITE_THE_PIPELINE
import dev.monosmith.jetbrains.words.SmithWordsConfig.say
import java.nio.file.Path

/**
 * Adds a ready-made pipeline (`smith pipeline`). The templates come from smith; a deploy template asks for
 * the app it builds. The plan is `--dry-run`, the files it would write; the file is the team's from then on.
 */
object AddPipelineUseCase {
    private val TITLE = say(JETBRAINS_PIPELINETEMPLATE_MONOSMITH_ADD_A_PIPELINE)

    /** Starts the flow in a background task. */
    fun start(
        ide: Project,
        root: Path,
    ) = BackgroundFlowClient.start(ide, TITLE) { run(IdePortsClient.of(ide, root)) }

    /**
     * @return true when smith wrote the pipeline
     */
    fun run(ports: FlowPortsContract): Boolean {
        val listed = ports.smith.run(listOf("pipeline"))
        if (!listed.ok) {
            ports.notice.warn(say(JETBRAINS_PIPELINETEMPLATE_NO_PIPELINE_WAS_ADDED, FailureTextMapper.describe(listed).toString()))
            return false
        }
        val templates = PipelineTemplatesMapper.toTemplates(listed.stdout)
        if (templates.isEmpty()) {
            ports.notice.warn(say(JETBRAINS_PIPELINETEMPLATE_SMITH_LISTED_NO_PIPELINE_UPDATE))
            return false
        }
        val rows = PipelineTemplatesMapper.rows(templates)
        val name = ports.prompt.pickOne(TITLE, say(JETBRAINS_PIPELINETEMPLATE_PICK_A_PIPELINE_IT_IS), rows)?.name ?: return false
        val template = templates.first { it.name == name }
        val app =
            if (template.needsApp) {
                ports.prompt.text(TITLE, say(JETBRAINS_PIPELINETEMPLATE_THE_APP_THIS_PIPELINE_BUILDS)) {
                    if (it.isEmpty()) say(JETBRAINS_PIPELINETEMPLATE_NAME_THE_APP) else null
                } ?: return false
            } else {
                null
            }
        val args = PipelineTemplatesMapper.arguments(template, app)
        return PlanThenApplyUseCase.run(
            ports,
            PlanStepContract(
                title = "$TITLE: ${template.name}?",
                planArgs = args + "--dry-run",
                header =
                    say(JETBRAINS_PIPELINETEMPLATE_THE_FILE_IS_YOURS_FROM),
                okText = say(JETBRAINS_PIPELINETEMPLATE_WRITE_THE_PIPELINE),
                offerCommit = false,
                apply = { args + "--yes" },
                done = say(JETBRAINS_PIPELINETEMPLATE_THE_PIPELINE_IS_WRITTEN, template.name.toString()),
            ),
        ) {}
    }
}
