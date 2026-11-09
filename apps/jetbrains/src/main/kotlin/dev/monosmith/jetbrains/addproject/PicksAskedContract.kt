package dev.monosmith.jetbrains.addproject

/**
 * What the check of the picks is asked about.
 *
 * @property language the language of the new project
 * @property picks the names ticked
 * @property outcomeId what is being built (`web`, `function`...); null on the long way
 * @property role the kind of project: `app`, `lib` or `package`
 */
data class PicksAskedContract(
    val language: String,
    val picks: List<String>,
    val outcomeId: String?,
    val role: String?,
)
