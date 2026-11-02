package dev.monosmith.jetbrains.prompt

/**
 * What the person answered in a plan dialog.
 *
 * @property commit the "Commit this change so it can be undone" box was ticked; it starts unticked
 * @property checks the state of each extra box, in the order the flow gave them
 * @property other the index of the extra button that closed the dialog, or -1 when the main button did
 */
data class PlanAnswerContract(
    val commit: Boolean,
    val checks: List<Boolean>,
    val other: Int,
)
