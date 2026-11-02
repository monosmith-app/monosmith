package dev.monosmith.jetbrains.convertfolder

import dev.monosmith.jetbrains.prompt.ChoiceRowContract

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

/** The command lines and the questions of `smith convert`, as its help documents them (smith has no JSON for them). */
object ConvertArgumentsAlgorithm {
    /** The two levels; the conservative one changes the least and is first. */
    val levelRows =
        listOf(
            ChoiceRowContract(
                "conservative",
                "Conservative",
                "(default)",
                "Keep the build and lint setup that works, and list what needs manual work.",
            ),
            ChoiceRowContract(
                "full",
                "Full",
                "",
                "Make smith the only tool. The old tooling is deleted after the plan is shown and you say yes.",
            ),
        )

    /** The lint choice: smith's ESLint config or the current setup. Keep is first. */
    val lintRows =
        listOf(
            ChoiceRowContract("keep", "Keep the current setup", "(default)", "Nothing about lint changes."),
            ChoiceRowContract("smith", "Use smith's ESLint config", "", "The old lint configs are deleted."),
        )

    /** The test runner choice. Keep is first. */
    val testRows =
        listOf(
            ChoiceRowContract("keep", "Keep the current setup", "(default)", "Nothing about tests changes."),
            ChoiceRowContract("jest", "Jest", "", "The other runner's configs and packages go."),
            ChoiceRowContract("vitest", "Vitest", "", "The other runner's configs and packages go."),
        )

    /** The build choice: smith's stack build or the current build setup. */
    val buildRows =
        listOf(
            ChoiceRowContract("smith", "Build with smith's stack build", "(default)", "smith runs the build."),
            ChoiceRowContract("keep", "Keep the current build setup", "", "Nothing about the build changes."),
        )

    /** The pipeline choice: the one smith generates or the current one. */
    val ciRows =
        listOf(
            ChoiceRowContract(
                "smith",
                "Use the pipeline smith generates",
                "(default)",
                "The pipelines are rewritten; the plan shows a diff of each.",
            ),
            ChoiceRowContract("keep", "Keep the current pipeline", "", "Nothing about CI changes."),
        )

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
