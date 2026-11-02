package dev.monosmith.jetbrains.health

/**
 * The change that clears a failed check, as `smith doctor --json` carries it.
 *
 * @property description the name `smith doctor --fix --only` takes
 * @property safe true when `smith doctor --fix --yes` alone may apply it
 * @property questions the questions of the fix; each has the labels of its options, the suggested one first
 */
data class DoctorFixContract(
    val description: String,
    val safe: Boolean,
    val questions: List<DoctorQuestionContract>,
)

/** One question of a fix: the person picks one answer. */
data class DoctorQuestionContract(
    val question: String,
    val options: List<String>,
)

/**
 * One check of `smith doctor --json`.
 *
 * @property warning a failed check that only warns
 * @property fix the fix smith can apply after the person agrees, or null
 */
data class DoctorFindingContract(
    val check: String,
    val ok: Boolean,
    val detail: String,
    val remedy: String,
    val warning: Boolean,
    val fix: DoctorFixContract?,
)

/** What a doctor report comes to. */
data class DoctorSummaryContract(
    val passed: Int,
    val warnings: Int,
    val failed: Int,
    val lines: List<String>,
)
