package dev.monosmith.jetbrains.dependencies

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/**
 * One dependency with a newer version, as `smith up --json` lists it.
 *
 * @property level `patch`, `minor` or `major`: how risky the update is
 * @property projects the projects that declare it
 */
data class UpdateContract(
    val plugin: String,
    val name: String,
    val current: String,
    val latest: String,
    val level: String,
    val projects: List<String>,
)

/** A tool (a language runtime) with a newer version than the workspace pins. */
data class ToolUpdateContract(
    val name: String,
    val pinned: String,
    val latest: String,
)

/** What `smith up --json` prints: it never writes. */
data class UpdatesReportContract(
    val updates: List<UpdateContract>,
    val tools: List<ToolUpdateContract>,
)

/** Reads `smith up --json`, groups the updates by risk and words the plan and the command lines. */
object UpdatesJsonMapper {
    /** The risk levels, safest first. */
    val levels = listOf("patch", "minor", "major")

    /**
     * @param json what smith printed
     * @return the updates and the newer tools
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun toReport(json: String): UpdatesReportContract {
        val root = JsonFieldsMapper.parseObject(json)
        return UpdatesReportContract(
            updates =
                JsonFieldsMapper
                    .records(root, "updates")
                    .map {
                        UpdateContract(
                            JsonFieldsMapper.text(it, "plugin"),
                            JsonFieldsMapper.text(it, "name"),
                            JsonFieldsMapper.text(it, "current"),
                            JsonFieldsMapper.text(it, "latest"),
                            JsonFieldsMapper.text(it, "level"),
                            JsonFieldsMapper.records(it, "sites").map { site -> JsonFieldsMapper.text(site, "project") }.distinct(),
                        )
                    }.filter { it.name.isNotEmpty() },
            tools =
                JsonFieldsMapper
                    .records(root, "tools")
                    .filter { JsonFieldsMapper.flag(it, "newer") }
                    .map {
                        ToolUpdateContract(
                            JsonFieldsMapper.text(it, "name"),
                            JsonFieldsMapper.text(it, "pinned"),
                            JsonFieldsMapper.text(it, "latest"),
                        )
                    },
        )
    }

    /** The updates of each level, the riskiest group first; a level with none is left out. */
    fun byRisk(updates: List<UpdateContract>): List<Pair<String, List<UpdateContract>>> =
        levels.reversed().mapNotNull { level -> updates.filter { it.level == level }.takeIf { it.isNotEmpty() }?.let { level to it } }

    /** What a level means, in a few words. */
    fun levelWords(level: String): String =
        when (level) {
            "major" -> "Major: may break your code; read the changelog"
            "minor" -> "Minor: new features, no breaking change"
            else -> "Patch: fixes only"
        }

    /** One line for an update: `zod 3.22.0 → 3.23.1 (3 projects)`. */
    fun line(update: UpdateContract): String =
        "${update.name} ${update.current} → ${update.latest} (${update.projects.size} ${if (update.projects.size == 1) "project" else "projects"})"

    /** The plan of an update, in words: each line says what moves and where. */
    fun describe(updates: List<UpdateContract>): String =
        updates.joinToString("\n") { "${line(it)} · ${it.level}\n    in ${it.projects.joinToString(", ")}" }

    /** The lines of "which updates?": the ones of the riskiest level are marked in the description; nothing risky is ticked at first. */
    fun rows(updates: List<UpdateContract>): List<ChoiceRowContract> =
        updates
            .sortedBy { levels.indexOf(it.level) }
            .map {
                ChoiceRowContract(
                    it.name,
                    "${it.name} ${it.current} → ${it.latest}",
                    it.level,
                    it.projects.joinToString(", "),
                    picked =
                        it.level != "major",
                )
            }

    /**
     * The arguments of `smith up` for the packages the person picked.
     *
     * @param names the picked packages, or empty for every patch and minor update
     * @param major true when a picked package is a major update, so `--major` lets it through
     */
    fun arguments(
        names: List<String>,
        major: Boolean,
    ): List<String> = listOf("up") + names + (if (major) listOf("--major") else emptyList()) + "--yes"

    /** The arguments of `smith up --align` for a rule; [apply] false is the plan. */
    fun alignArguments(
        rule: String,
        apply: Boolean,
    ): List<String> = listOf("up", "--align", "--prefer", rule, if (apply) "--yes" else "--dry-run")
}
