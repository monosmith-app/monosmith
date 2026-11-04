package dev.monosmith.jetbrains.addchoices

import dev.monosmith.jetbrains.flowports.PromptPort
import dev.monosmith.jetbrains.prompt.ChoiceRowContract

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
     * @param role the kind of project (`app`, `lib`, `package`): the lists show only the choices that fit it
     * @param check checks the picks with smith before the lists are left: what is wrong in plain words, or an empty text;
     *   a problem is said and the questions start again
     * @return the names for `--with`, or null when the person cancelled
     */
    fun pick(
        prompt: PromptPort,
        choices: List<ChoiceOptionContract>,
        language: String,
        preset: List<String>,
        outcomeId: String?,
        ofLayers: Map<String, String>,
        about: Map<String, String> = emptyMap(),
        role: String? = null,
        check: (List<String>) -> String = { "" },
    ): List<String>? {
        while (true) {
            val picked = askAxes(prompt, choices, language, preset, outcomeId, ofLayers, about, role) ?: return null
            val problem = check(picked)
            if (problem.isEmpty()) return picked
            val again =
                prompt.pickOne(
                    "MonoSmith: Add project: check",
                    problem,
                    listOf(
                        ChoiceRowContract(AGAIN, "Pick again", "", "Go back to the questions."),
                        ChoiceRowContract(CANCEL, "Cancel", "", "Add nothing."),
                    ),
                    "Choose",
                )
            if (again?.name != AGAIN) return null
        }
    }

    private const val AGAIN = "again"
    private const val CANCEL = "cancel"

    @Suppress("LongParameterList")
    private fun askAxes(
        prompt: PromptPort,
        choices: List<ChoiceOptionContract>,
        language: String,
        preset: List<String>,
        outcomeId: String?,
        ofLayers: Map<String, String>,
        about: Map<String, String>,
        role: String?,
    ): List<String>? {
        val axes = ChoiceRowsAlgorithm.byAxis(choices, language, role)
        val picked = preset.toMutableList()
        val settled = axes.filter { axis -> axis.choices.any { it.name in preset } }.map { it.axis }.toSet()
        val asked = ChoiceRowsAlgorithm.askedAxes(choices, language)

        fun chosen() = choices.filter { it.name in picked }
        for (axis in axes) {
            if (axis.axis in settled) continue
            val layer = (ofLayers[axis.axis] ?: axis.axis).lowercase()
            val title = "MonoSmith: Add project: $layer"
            val gloss = GlossaryAlgorithm.line(axis.axis, ofLayers, about)
            if (axis.axis in asked) {
                val needing = chosen().filter { axis.axis in it.needs }
                if (needing.isEmpty()) continue
                val header =
                    GlossaryAlgorithm.withLine(
                        "${needing.joinToString(", ") { it.label }} needs an answer on the $layer axis. smith never picks it. " +
                            "The suggested answer is first.",
                        gloss,
                    )
                val rows = ChoiceRowsAlgorithm.askedRows(axis.choices, chosen())
                picked.add((prompt.pickOne(title, header, rows) ?: return null).name)
            } else if (axis.many) {
                val rows = ChoiceRowsAlgorithm.manyChoiceRows(axis, chosen(), ofLayers)
                picked.addAll(
                    prompt.pickMany(title, GlossaryAlgorithm.withLine("Tick what the project uses. The defaults are ticked.", gloss), rows)
                        ?: return null,
                )
            } else {
                val rows = ChoiceRowsAlgorithm.singleChoiceRows(axis, chosen(), outcomeId, ofLayers)
                val answer =
                    prompt.pickOne(title, GlossaryAlgorithm.withLine("Pick one $layer. The first answer is the default.", gloss), rows)
                        ?: return null
                if (answer.name.isNotEmpty()) picked.add(answer.name)
            }
        }
        return picked
    }
}
