package dev.monosmith.jetbrains.addchoices

import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/** The questions of the choices of a new project, built from the catalogue of `smith choices --json`. No name is known here. */
object ChoiceRowsAlgorithm {
    /**
     * The choices a project of [language] can be built from, grouped by axis.
     *
     * @return the axes in the order smith lists them, each with its choices
     */
    fun byAxis(
        choices: List<ChoiceOptionContract>,
        language: String,
    ): List<AxisGroupContract> =
        choices
            .filter { it.language.isEmpty() || it.language == language }
            .groupBy { it.axis }
            .map { (axis, list) -> AxisGroupContract(axis, list.any { it.many }, list) }

    /**
     * The axes of a language that a choice needs and that have no default: the trigger and the runtime
     * of a function app. smith never picks them, so the flow asks each once a choice that needs it is picked.
     */
    fun askedAxes(
        choices: List<ChoiceOptionContract>,
        language: String,
    ): List<String> {
        val axes = byAxis(choices, language)
        val defaulted = axes.filter { axis -> axis.choices.any { it.isDefault } }.map { it.axis }.toSet()
        val listed = axes.map { it.axis }.toSet()
        return axes.flatMap { axis -> axis.choices.flatMap { it.needs } }.distinct().filter { it in listed && it !in defaulted }
    }

    /**
     * Whether a choice can be offered for what is being built. A cloud host fits only a function app, and
     * a desktop shell that sits among the frameworks fits only a desktop app.
     *
     * @param outcomeId what is being built (`web`, `function`, `desktop`); null on the long way, which offers all
     */
    fun fitsOutcome(
        choice: ChoiceOptionContract,
        outcomeId: String?,
    ): Boolean {
        if (choice.layer == "host") return outcomeId == null || outcomeId == "function"
        val wrapsAnApp = choice.layer == "shell" && choice.axis != "shell"
        return !wrapsAnApp || outcomeId == null || outcomeId == "desktop"
    }

    private fun goesWith(
        choice: ChoiceOptionContract,
        picked: List<ChoiceOptionContract>,
    ): Boolean = picked.all { it.name !in choice.conflicts && choice.name !in it.conflicts }

    /**
     * The lines of a single-choice list, with the answer that Enter takes first: the default of the axis
     * when it has one, else "None" (an axis with no default adds nothing when it is left out). A choice
     * that does not go with what is picked, or with what is being built, is not listed.
     *
     * @param ofLayers the title of each layer, to word "None"
     */
    fun singleChoiceRows(
        axis: AxisGroupContract,
        picked: List<ChoiceOptionContract>,
        outcomeId: String?,
        ofLayers: Map<String, String> = emptyMap(),
    ): List<ChoiceRowContract> {
        val fits = axis.choices.filter { fitsOutcome(it, outcomeId) && goesWith(it, picked) }
        val ordered = fits.sortedByDescending { it.isDefault }
        val rows =
            ordered.map {
                ChoiceRowContract(
                    name = it.name,
                    label = it.label,
                    description =
                        listOf(
                            if (it.isDefault) "(default)" else "",
                            if (it.requires.isEmpty()) "" else "needs ${it.requires.joinToString(", ")}",
                        ).filter(String::isNotEmpty)
                            .joinToString(" "),
                    detail = it.description,
                )
            }
        if (fits.any { it.isDefault }) return rows
        val layer = (ofLayers[axis.axis] ?: axis.axis).lowercase()
        return listOf(ChoiceRowContract("", "None", "(default) nothing extra", "Leave out the $layer: you can add one later.")) + rows
    }

    /**
     * The lines of the many-answers question of an axis; the defaults start ticked.
     *
     * @param ofLayers the title of each layer
     */
    fun manyChoiceRows(
        axis: AxisGroupContract,
        picked: List<ChoiceOptionContract>,
        ofLayers: Map<String, String> = emptyMap(),
    ): List<ChoiceRowContract> =
        axis.choices.filter { goesWith(it, picked) }.map {
            val layer = if (it.layer.isEmpty()) "" else ofLayers[it.layer] ?: it.layer
            ChoiceRowContract(
                name = it.name,
                label = it.label,
                description =
                    listOf(
                        if (it.isDefault) "(default)" else "",
                        layer,
                        if (it.requires.isEmpty()) "" else "needs ${it.requires.joinToString(", ")}",
                    ).filter(String::isNotEmpty)
                        .joinToString(" "),
                detail = it.description,
                picked = it.isDefault,
            )
        }

    /**
     * The lines of the question about an axis that smith never picks (a trigger, a runtime): the choices
     * that go with what is picked, the suggested one first.
     */
    fun askedRows(
        choices: List<ChoiceOptionContract>,
        picked: List<ChoiceOptionContract>,
    ): List<ChoiceRowContract> =
        choices
            .filter { goesWith(it, picked) }
            .sortedByDescending { it.suggested }
            .map { ChoiceRowContract(it.name, it.name, if (it.suggested) "(suggested)" else "", it.description) }

    /**
     * Says the choices of a request in words, and which of them the catalogue marks as the default, so a
     * person who took every first answer can see what they got.
     *
     * @param named the names that go to `--with`
     * @param stack the stack of the request, or null
     */
    fun inWords(
        named: List<String>,
        stack: String?,
        language: String,
        choices: List<ChoiceOptionContract>,
        ofLayers: Map<String, String> = emptyMap(),
    ): String {
        if (stack != null) return "Built from the stack $stack."
        val axes = byAxis(choices, language)
        val chosen = named.mapNotNull { name -> axes.flatMap { it.choices }.firstOrNull { it.name == name } }
        if (chosen.isEmpty()) return "Built with the default choices of $language: nothing extra was chosen."
        val asked = askedAxes(choices, language)
        val taken = chosen.map { it.axis }.toSet()
        val without =
            axes
                .filter { axis -> !axis.many && axis.axis !in asked && axis.axis !in taken && axis.choices.none { it.isDefault } }
                .map { "no ${(ofLayers[it.axis] ?: it.axis).lowercase()}" }
        val list = chosen.joinToString(", ") { "${it.label}${if (it.isDefault) " (default)" else ""}" }
        return "Your choices: $list${if (without.isEmpty()) "" else "; ${without.joinToString(", ")}"}."
    }
}
