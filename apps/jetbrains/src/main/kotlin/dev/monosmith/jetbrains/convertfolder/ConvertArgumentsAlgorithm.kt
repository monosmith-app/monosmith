package dev.monosmith.jetbrains.convertfolder

/**
 * What the person decided about a conversion. The four choices apply to the full level only.
 *
 * @property level `conservative` (keep what works, list the rest) or `full` (make smith the only tool)
 */
data class ConvertChoicesContract(
    val level: String = "conservative",
    val lint: String = "keep",
    val test: String = "keep",
    val build: String = "smith",
    val ci: String = "smith",
)

/** The command lines of `smith convert`. Its questions come from `smith convert --questions --json`. */
object ConvertArgumentsAlgorithm {
    /** The arguments of `smith convert`; [apply] false is the plan (`--dry-run`), true writes (`--yes`). */
    fun arguments(
        choices: ConvertChoicesContract,
        apply: Boolean,
    ): List<String> =
        buildList {
            addAll(listOf("convert", "--level", choices.level))
            if (choices.level ==
                "full"
            ) {
                addAll(listOf("--lint", choices.lint, "--test", choices.test, "--build", choices.build, "--ci", choices.ci))
            }
            add(if (apply) "--yes" else "--dry-run")
        }
}
