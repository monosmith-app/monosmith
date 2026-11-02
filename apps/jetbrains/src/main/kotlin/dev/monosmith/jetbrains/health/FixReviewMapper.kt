package dev.monosmith.jetbrains.health

/** Reads what `smith doctor --fix --only` prints. */
object FixReviewMapper {
    /** The arguments of the fix of one check: `doctor --fix --only <fix> [--choose 1,2] [--yes]`. [answers] are 1-based. */
    fun arguments(
        fix: DoctorFixContract,
        answers: List<Int>,
        apply: Boolean,
    ): List<String> =
        listOf("doctor", "--fix", "--only", fix.description) +
            (if (answers.isEmpty()) emptyList() else listOf("--choose", answers.joinToString(","))) +
            (if (apply) listOf("--yes") else emptyList())

    /**
     * The review of a fix: the diffs, the commands and the plan each command prints. The review of an
     * unsafe fix ends non-zero when nobody could answer; it is still the review. What follows "not applied"
     * tells a terminal to run a command: here the Apply button is that command.
     */
    fun review(stdout: String): String = stdout.split(Regex("\\r?\\n(?=not applied:)")).first().trim()

    /** True when doctor still fails on another check after the fix ("after fixing ..." in its output). */
    fun stillFailing(
        code: Int,
        lines: List<String>,
    ): Boolean = code != 0 && lines.any { it.startsWith("after fixing") }
}
