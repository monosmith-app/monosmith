package dev.monosmith.jetbrains.swapchoice

import dev.monosmith.jetbrains.addchoices.AxisGroupContract
import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/** The command lines and the lists of swapping one choice of a project (its test runner, its style). */
object SwapArgumentsAlgorithm {
    /** The arguments of `smith swap`; [apply] false makes the plan (`--dry-run`), true writes (`--yes`). */
    fun arguments(
        project: String,
        axis: String,
        choice: String,
        apply: Boolean,
    ): List<String> = listOf("swap", project, axis, choice, if (apply) "--yes" else "--dry-run")

    /** The lines of "which decision?": one for each axis the language offers, with how many answers it has. */
    fun axisRows(
        axes: List<AxisGroupContract>,
        ofLayers: Map<String, String>,
    ): List<ChoiceRowContract> =
        axes.map {
            ChoiceRowContract(
                it.axis,
                ofLayers[it.axis] ?: it.axis,
                "${it.choices.size} answers",
                if (it.many) "Several can be used at once." else "One answer.",
            )
        }

    /** The lines of "which answer?": the default is marked. */
    fun choiceRows(axis: AxisGroupContract): List<ChoiceRowContract> =
        axis.choices.map { ChoiceRowContract(it.name, it.label, if (it.isDefault) "(default)" else "", it.description) }
}
