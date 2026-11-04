package dev.monosmith.jetbrains.pipelinetemplate

import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.smithcommand.SmithFlagsConfig

/**
 * One ready-made pipeline of `smith pipeline`.
 *
 * @property name the name `smith pipeline <name>` takes
 * @property providers the CI systems it exists for, as smith prints them
 * @property description what the pipeline does
 * @property needsApp a deploy pipeline builds an app, so `--project` is required
 */
data class PipelineTemplateContract(
    val name: String,
    val providers: String,
    val description: String,
    val needsApp: Boolean,
)

/** Reads the list that `smith pipeline` prints: one template on each line, its columns apart by two spaces. */
object PipelineTemplatesMapper {
    private val name = Regex("^[a-z][a-z0-9-]*")
    private val gap = Regex("\\s{2,}")

    /**
     * @param text what smith printed
     * @return the templates, in smith's order; a line that does not start with a name is skipped
     */
    fun toTemplates(text: String): List<PipelineTemplateContract> =
        text.lines().mapNotNull { line ->
            val cells = line.trim().split(gap)
            val first = cells.firstOrNull().orEmpty()
            val found = name.find(first)?.value ?: return@mapNotNull null
            PipelineTemplateContract(
                found,
                cells.getOrElse(1) { "" },
                cells.drop(2).joinToString(" "),
                first.contains(SmithFlagsConfig.PROJECT_LONG),
            )
        }

    /** One row for each template: the CI systems, then what it does, and a note when it asks for an app. */
    fun rows(templates: List<PipelineTemplateContract>): List<ChoiceRowContract> =
        templates.map {
            ChoiceRowContract(
                it.name,
                it.name,
                if (it.providers.isEmpty()) "" else "[${it.providers}]",
                it.description + if (it.needsApp) " (asks for the app)" else "",
            )
        }

    /** The arguments after `smith`: the template and, for a deploy pipeline, the app. */
    fun arguments(
        template: PipelineTemplateContract,
        app: String?,
    ): List<String> =
        listOf("pipeline", template.name) + if (app.isNullOrEmpty()) emptyList() else listOf(SmithFlagsConfig.PROJECT_LONG, app)
}
