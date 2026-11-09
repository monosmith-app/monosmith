package dev.monosmith.jetbrains.planapply

import dev.monosmith.jetbrains.prompt.PlanAnswerContract

/**
 * One action with a plan: smith shows what it would do, the person answers, and only then smith does it.
 *
 * @property title the title of the dialog and of the progress bar
 * @property planArgs the arguments of the plan (`--dry-run`), or null when smith has no plan command
 * @property fixedPlan the text to show when [planArgs] is null
 * @property header the sentence above the plan: what the answer decides and that nothing is changed yet
 * @property okText the verb of the main button
 * @property offerCommit show the unticked "Commit this change so it can be undone" box; false when the command has no `--commit`
 * @property others the label of each extra button (a second way to answer)
 * @property apply the arguments of the change for the answer the person gave; null leaves the action
 * @property done the sentence for a success, such as "web is removed"
 * @property operation the words that name the operation for the Commit button, such as "remove web"
 */
data class PlanStepContract(
    val title: String,
    val planArgs: List<String>?,
    val fixedPlan: String = "",
    val header: String,
    val okText: String,
    val offerCommit: Boolean,
    val others: List<String> = emptyList(),
    val apply: (PlanAnswerContract) -> List<String>?,
    val done: String,
    val operation: String = "",
)
