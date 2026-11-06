package dev.monosmith.jetbrains.releaseflow

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_A_BREAKING_CHANGE_1_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_A_FIX_1_2_3
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_A_NEW_FEATURE_1_2
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_CHANGE_ITS_BUMP_A_REQUEST
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_HOLD_IT_BACK_IT_DOES
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_KEEP_THE_BUMPS_AS_THEY
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT_THE
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_RELEASE_IT_ANYWAY_THE_SAME
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_RELEASEFLOW_THE_COMMITS_DECIDE_THE_LEVELS
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

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
            ChoiceRowContract("patch", "patch", say(JETBRAINS_ADDCHOICES_DEFAULT), say(JETBRAINS_RELEASEFLOW_A_FIX_1_2_3)),
            ChoiceRowContract("minor", "minor", "", say(JETBRAINS_RELEASEFLOW_A_NEW_FEATURE_1_2)),
            ChoiceRowContract("major", "major", "", say(JETBRAINS_RELEASEFLOW_A_BREAKING_CHANGE_1_2)),
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
        return say(JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT_THE, names.toString())
    }

    /** The lines of "hold packages back?": nothing is ticked at first. */
    fun holdRows(packages: List<PlannedPackageContract>): List<ChoiceRowContract> =
        packages.map {
            ChoiceRowContract(
                it.name,
                it.name,
                "${it.current.ifEmpty { "new" }} → ${it.next} · ${it.level}",
                say(JETBRAINS_RELEASEFLOW_HOLD_IT_BACK_IT_DOES),
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
                say(JETBRAINS_RELEASEFLOW_KEEP_THE_BUMPS_AS_THEY),
                say(JETBRAINS_ADDCHOICES_DEFAULT),
                if (changed.isEmpty()) say(JETBRAINS_RELEASEFLOW_THE_COMMITS_DECIDE_THE_LEVELS) else "${changed.size} changed.",
            ),
        ) +
            packages.filter { it.current.isNotEmpty() }.map {
                val level = changed[it.name] ?: it.level
                ChoiceRowContract(
                    it.name,
                    it.name,
                    "${it.current} → ${it.next} · $level${if (it.name in changed) " (changed)" else ""}",
                    say(JETBRAINS_RELEASEFLOW_CHANGE_ITS_BUMP_A_REQUEST),
                )
            }

    /** The lines of "release a package without a releasing commit?": a package is released only if the person ticks it. */
    fun forcedRows(unchanged: List<String>): List<ChoiceRowContract> =
        unchanged.map {
            ChoiceRowContract(
                it,
                it,
                say(JETBRAINS_RELEASEFLOW_NO_COMMIT_RELEASES_IT),
                say(JETBRAINS_RELEASEFLOW_RELEASE_IT_ANYWAY_THE_SAME),
            )
        }
}
