package dev.monosmith.jetbrains.swapchoice

import dev.monosmith.jetbrains.addchoices.AxisGroupContract
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_ONE_ANSWER
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SWAPCHOICE_SEVERAL_CAN_BE_USED_AT
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

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
                if (it.many) say(JETBRAINS_SWAPCHOICE_SEVERAL_CAN_BE_USED_AT) else say(JETBRAINS_SWAPCHOICE_ONE_ANSWER),
            )
        }

    /** The lines of "which answer?": the default is marked. */
    fun choiceRows(axis: AxisGroupContract): List<ChoiceRowContract> =
        axis.choices.map {
            ChoiceRowContract(
                it.name,
                it.label,
                if (it.isDefault) say(JETBRAINS_ADDCHOICES_DEFAULT) else "",
                it.description,
            )
        }
}
