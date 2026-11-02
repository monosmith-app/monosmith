package dev.monosmith.jetbrains.addchoices

import com.intellij.openapi.project.Project
import dev.monosmith.jetbrains.prompt.PromptClient

/**
 * Asks the choices of a new project, axis by axis. What the first questions settled is kept and not
 * asked again (a website's framework, a function's cloud). A list of one answer starts on its default,
 * or on "None" when the axis has no default, so taking every first answer never adds a shell or a style.
 * Blocks: call it from a background flow.
 */
object PickChoicesUseCase {
    /**
     * @param choices every choice smith lists
     * @param language the language of the new project
     * @param preset the choices the first questions named (a framework, a desktop shell, a cloud)
     * @param outcomeId what is being built; null on the long way
     * @param ofLayers the title of each layer
     * @return the names for `--with`, or null when the person cancelled
     */
    fun pick(
        ide: Project,
        choices: List<ChoiceOptionContract>,
        language: String,
        preset: List<String>,
        outcomeId: String?,
        ofLayers: Map<String, String>,
    ): List<String>? {
        val axes = ChoiceRowsAlgorithm.byAxis(choices, language)
        val picked = preset.toMutableList()
        val settled = axes.filter { axis -> axis.choices.any { it.name in preset } }.map { it.axis }.toSet()
        val asked = ChoiceRowsAlgorithm.askedAxes(choices, language)

        fun chosen() = choices.filter { it.name in picked }
        for (axis in axes) {
            if (axis.axis in settled) continue
            val layer = (ofLayers[axis.axis] ?: axis.axis).lowercase()
            val title = "MonoSmith: Add project: $layer"
            if (axis.axis in asked) {
                val needing = chosen().filter { axis.axis in it.needs }
                if (needing.isEmpty()) continue
                val header =
                    "${needing.joinToString(", ") { it.label }} needs an answer on the $layer axis. smith never picks it. " +
                        "The suggested answer is first."
                val rows = ChoiceRowsAlgorithm.askedRows(axis.choices, chosen())
                picked.add((PromptClient.pickOne(ide, title, header, rows) ?: return null).name)
            } else if (axis.many) {
                val rows = ChoiceRowsAlgorithm.manyChoiceRows(axis, chosen(), ofLayers)
                picked.addAll(
                    PromptClient.pickMany(ide, title, "Tick what the project uses. The defaults are ticked.", rows) ?: return null,
                )
            } else {
                val rows = ChoiceRowsAlgorithm.singleChoiceRows(axis, chosen(), outcomeId, ofLayers)
                val answer = PromptClient.pickOne(ide, title, "Pick one $layer. The first answer is the default.", rows) ?: return null
                if (answer.name.isNotEmpty()) picked.add(answer.name)
            }
        }
        return picked
    }
}
