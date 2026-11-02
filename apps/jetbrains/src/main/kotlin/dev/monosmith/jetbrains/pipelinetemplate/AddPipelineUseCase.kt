package dev.monosmith.jetbrains.pipelinetemplate

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.flowports.FlowPortsContract
import dev.monosmith.jetbrains.flowports.IdePortsClient
import dev.monosmith.jetbrains.planapply.PlanStepContract
import dev.monosmith.jetbrains.planapply.PlanThenApplyUseCase
import dev.monosmith.jetbrains.smithrun.BackgroundFlowClient
import dev.monosmith.jetbrains.smithrun.FailureTextMapper
import java.nio.file.Path

/**
 * Adds a ready-made pipeline (`smith pipeline`). The templates come from smith; a deploy template asks for
 * the app it builds. The plan is `--dry-run`, the files it would write; the file is the team's from then on.
 */
object AddPipelineUseCase {
    private const val TITLE = "MonoSmith: add a pipeline"

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
            ports.notice.warn("No pipeline was added. ${FailureTextMapper.describe(listed)}")
            return false
        }
        val templates = PipelineTemplatesMapper.toTemplates(listed.stdout)
        if (templates.isEmpty()) {
            ports.notice.warn("smith listed no pipeline. Update smith.")
            return false
        }
        val rows = PipelineTemplatesMapper.rows(templates)
        val name = ports.prompt.pickOne(TITLE, "Pick a pipeline. It is written beside the generated CI.", rows)?.name ?: return false
        val template = templates.first { it.name == name }
        val app =
            if (template.needsApp) {
                ports.prompt.text(TITLE, "The app this pipeline builds (the name of a project):") {
                    if (it.isEmpty()) "Name the app." else null
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
                    "The file is yours from then on: smith upgrade never rewrites it. " +
                        "A file that exists is kept. Nothing was written yet.",
                okText = "Write the pipeline",
                offerCommit = false,
                apply = { args + "--yes" },
                done = "The ${template.name} pipeline is written",
            ),
        ) {}
    }
}
