package dev.monosmith.jetbrains.releaseflow

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/**
 * A package the release would publish.
 *
 * @property current the version now; empty for a package never released
 * @property next the version the release makes
 * @property level `patch`, `minor` or `major`
 */
data class PlannedPackageContract(
    val name: String,
    val next: String,
    val level: String,
    val current: String,
)

/**
 * What the person decided about a release.
 *
 * @property hold the packages that do not release this time
 * @property bump `<project>=patch|minor|major` for each bump the person changed
 * @property forced `<project>=patch|minor|major` for each package no commit releases but the person released anyway
 */
data class ReleaseChoicesContract(
    val hold: List<String> = emptyList(),
    val bump: List<String> = emptyList(),
    val forced: List<String> = emptyList(),
)

/** The command lines, the readers of `smith release --dry-run --json` and the questions of the release flow. */
object ReleaseArgumentsAlgorithm {
    /** The levels a bump can take, with what each means. */
    val levelRows =
        listOf(
            ChoiceRowContract("patch", "patch", "(default)", "A fix: 1.2.3 → 1.2.4."),
            ChoiceRowContract("minor", "minor", "", "A new feature: 1.2.3 → 1.3.0."),
            ChoiceRowContract("major", "major", "", "A breaking change: 1.2.3 → 2.0.0. On a 0.x package, major gives 1.0.0."),
        )

    /**
     * The packages of `smith release --dry-run --json`.
     *
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun planned(json: String): List<PlannedPackageContract> =
        JsonFieldsMapper
            .records(JsonFieldsMapper.parseObject(json), "results")
            .filter { JsonFieldsMapper.text(it, "project").isNotEmpty() }
            .map {
                PlannedPackageContract(
                    JsonFieldsMapper.text(it, "project"),
                    JsonFieldsMapper.text(it, "next"),
                    JsonFieldsMapper.text(it, "level"),
                    JsonFieldsMapper.text(it, "current"),
                )
            }

    /**
     * The packages no commit releases, from `smith release --dry-run --json`.
     *
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun unchanged(json: String): List<String> =
        JsonFieldsMapper.words(JsonFieldsMapper.parseObject(json), "unchanged").filter { it.isNotEmpty() }

    /** The arguments of `smith release`; [publish] false makes the plan (`--dry-run`), true publishes (`--yes`). */
    fun arguments(
        choices: ReleaseChoicesContract,
        publish: Boolean,
    ): List<String> =
        buildList {
            addAll(listOf("release", if (publish) "--yes" else "--dry-run"))
            if (choices.hold.isNotEmpty()) addAll(listOf("--hold", choices.hold.joinToString(",")))
            (choices.bump + choices.forced).forEach { addAll(listOf("--bump", it)) }
            if (choices.forced.isNotEmpty()) add("--force")
        }

    /** Says in plain words what a release without a releasing commit does. */
    fun forcedConsequences(forced: List<String>): String {
        if (forced.isEmpty()) return ""
        val names = forced.joinToString(", ") { it.substringBefore("=") }
        return "\n\n$names: no commit releases it. The new tag points at HEAD, and the same code is published " +
            "under a new version, with a GitHub Release and the assets."
    }

    /** The lines of "hold packages back?": nothing is ticked at first. */
    fun holdRows(packages: List<PlannedPackageContract>): List<ChoiceRowContract> =
        packages.map {
            ChoiceRowContract(
                it.name,
                it.name,
                "${it.current.ifEmpty { "new" }} → ${it.next} · ${it.level}",
                "Hold it back: it does not release this time. A package that releases only because of it is held back too.",
            )
        }

    /** The lines of "change a bump?": the first row keeps them; a package that was never released has no bump to change. */
    fun bumpRows(
        packages: List<PlannedPackageContract>,
        changed: Map<String, String>,
    ): List<ChoiceRowContract> =
        listOf(
            ChoiceRowContract(
                "",
                "Keep the bumps as they are",
                "(default)",
                if (changed.isEmpty()) "The commits decide the levels." else "${changed.size} changed.",
            ),
        ) +
            packages.filter { it.current.isNotEmpty() }.map {
                val level = changed[it.name] ?: it.level
                ChoiceRowContract(
                    it.name,
                    it.name,
                    "${it.current} → ${it.next} · $level${if (it.name in changed) " (changed)" else ""}",
                    "Change its bump: a request, not an inference.",
                )
            }

    /** The lines of "release a package without a releasing commit?": a package is released only if the person ticks it. */
    fun forcedRows(unchanged: List<String>): List<ChoiceRowContract> =
        unchanged.map {
            ChoiceRowContract(
                it,
                it,
                "no commit releases it",
                "Release it anyway: the same code under a new version. It creates a tag at HEAD, publishes to the registry, makes a GitHub Release and cannot be undone.",
            )
        }
}
