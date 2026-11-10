package dev.monosmith.jetbrains.prompt

/**
 * One line of a question: an answer the person can pick.
 *
 * @property name the value the flow works with; empty for the line that leaves an axis out
 * @property label what the person reads
 * @property description a short qualifier after the label, such as `(default)`
 * @property detail one sentence that says what the answer means
 * @property picked for a list of several answers: ticked at first (the defaults)
 * @property search more words that find the row in the filter box (aliases, frameworks)
 */
data class ChoiceRowContract(
    val name: String,
    val label: String,
    val description: String = "",
    val detail: String = "",
    val picked: Boolean = false,
    val search: String = "",
)
