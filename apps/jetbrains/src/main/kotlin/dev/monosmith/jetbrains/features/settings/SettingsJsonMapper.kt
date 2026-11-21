package dev.monosmith.jetbrains.features.settings

import dev.monosmith.jetbrains.shared.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.shared.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.JETBRAINS_SETTINGS_WHY_THE_DEFAULT
import dev.monosmith.jetbrains.shared.words.SmithWordsConfig.say

/**
 * One setting of smith.toml, as `smith config list --json` lists it.
 *
 * @property type `string`, `boolean`, `integer`, `array`...
 * @property reason why the default is what it is; empty when smith says none
 * @property value the value written in smith.toml, as JSON text; empty when the setting is not set
 * @property set true when smith.toml sets it
 * @property options the values the setting accepts when it has a fixed set of them (one word, or the words a list
 *   holds), each explained; empty for any other setting
 */
data class SettingContract(
    val path: String,
    val type: String,
    val description: String,
    val reason: String,
    val value: String,
    val set: Boolean,
    val options: List<SettingOptionContract>,
)

/**
 * One value of a setting with a fixed set of values.
 *
 * @property value the word written in smith.toml
 * @property description one line that says what the value does
 * @property default true for the value (or each value of a list) the setting takes when it is not set
 */
data class SettingOptionContract(
    val value: String,
    val description: String,
    val default: Boolean,
)

/**
 * A value of a project that its manifest owns: shown read-only, with the file that holds it.
 *
 * @property key name, version, language, kind or root
 * @property source the workspace-relative file that holds it
 */
data class ProjectFactContract(
    val key: String,
    val value: String,
    val description: String,
    val source: String,
)

/**
 * What `smith config list -p <project> --json` prints.
 *
 * @property manifest the file that owns the read-only values
 * @property facts the read-only values
 * @property settings the settings of `[projects.<name>]`, at `projects.<name>.<key>`
 */
data class ProjectViewContract(
    val name: String,
    val manifest: String,
    val facts: List<ProjectFactContract>,
    val settings: List<SettingContract>,
)

/** Reads the settings and words the questions and command lines of the settings editor. */
object SettingsJsonMapper {
    /**
     * @param json what smith printed
     * @return the settings that have a path
     * @throws IllegalStateException when the text is not a JSON list
     */
    fun toSettings(json: String): List<SettingContract> =
        JsonFieldsMapper
            .parseArray(json)
            .filter { it.isJsonObject }
            .map { it.asJsonObject }
            .map {
                SettingContract(
                    JsonFieldsMapper.text(it, "path"),
                    JsonFieldsMapper.text(it, "type"),
                    JsonFieldsMapper.text(it, "description"),
                    JsonFieldsMapper.text(it, "reason"),
                    JsonFieldsMapper.text(it, "value"),
                    JsonFieldsMapper.flag(it, "set"),
                    options(it),
                )
            }.filter { it.path.isNotEmpty() }

    private fun options(record: com.google.gson.JsonObject): List<SettingOptionContract> =
        JsonFieldsMapper
            .records(record, "options")
            .map {
                SettingOptionContract(
                    JsonFieldsMapper.text(it, "value"),
                    JsonFieldsMapper.text(it, "description"),
                    JsonFieldsMapper.flag(it, "default"),
                )
            }.filter { it.value.isNotEmpty() }

    /**
     * @param json what `smith config list -p <project> --json` printed
     * @return the view of the project
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun toProjectView(json: String): ProjectViewContract {
        val record = JsonFieldsMapper.parseObject(json)
        return ProjectViewContract(
            JsonFieldsMapper.text(record, "name"),
            JsonFieldsMapper.text(record, "manifest"),
            JsonFieldsMapper.records(record, "facts").map {
                ProjectFactContract(
                    JsonFieldsMapper.text(it, "path"),
                    JsonFieldsMapper.text(it, "value"),
                    JsonFieldsMapper.text(it, "description"),
                    JsonFieldsMapper.text(it, "source"),
                )
            },
            toSettings(record.get("settings")?.takeIf { it.isJsonArray }?.toString() ?: "[]"),
        )
    }

    /** The value as a person types it: `"main"` is `main`, `["a","b"]` is `a, b`. */
    fun shown(setting: SettingContract): String {
        val text = setting.value.trim()
        if (text.isEmpty()) return ""
        return runCatching {
            val element =
                com.google.gson.JsonParser
                    .parseString(text)
            when {
                element.isJsonArray -> element.asJsonArray.joinToString(", ") { if (it.isJsonPrimitive) it.asString else it.toString() }
                element.isJsonPrimitive -> element.asString
                else -> text
            }
        }.getOrDefault(text)
    }

    /** The lines of "which setting?": the value it has (or `default`) after the name, and what it is for. */
    fun rows(settings: List<SettingContract>): List<ChoiceRowContract> =
        settings.map {
            ChoiceRowContract(
                it.path,
                it.path,
                if (it.set) "= ${shown(it)}" else say(JETBRAINS_ADDCHOICES_DEFAULT),
                it.description,
            )
        }

    /**
     * The lines of a setting that takes one of a few values: each says what it does and the default is marked. For
     * one value the current one is first and marked; for a list each value the list holds now (or, when it is not
     * set, each default) starts ticked.
     */
    fun enumRows(setting: SettingContract): List<ChoiceRowContract> {
        val many = setting.type == "array"
        val current = shown(setting).split(",").map(String::trim).filter(String::isNotEmpty)
        val options = if (many) setting.options else setting.options.sortedByDescending { it.value in current }
        return options.map {
            ChoiceRowContract(
                it.value,
                it.value,
                listOfNotNull(
                    say(JETBRAINS_ADDCHOICES_DEFAULT).takeIf { _ -> it.default },
                    "(current)".takeIf { _ -> !many && it.value in current },
                ).joinToString(" "),
                it.description,
                picked = many && if (setting.set) it.value in current else it.default,
            )
        }
    }

    /** The two answers of a true-or-false setting; the current one is first. */
    fun booleanRows(setting: SettingContract): List<ChoiceRowContract> {
        val current = if (setting.set) shown(setting) else ""
        val rows =
            listOf(
                ChoiceRowContract(
                    "true",
                    "true",
                    if (current ==
                        "true"
                    ) {
                        "(current)"
                    } else {
                        ""
                    },
                ),
                ChoiceRowContract("false", "false", if (current == "false") "(current)" else ""),
            )
        return if (current == "false") rows.reversed() else rows
    }

    /** `smith config set <setting> <value...>`: an array is one value per item, split on commas. */
    fun setArguments(
        setting: SettingContract,
        text: String,
    ): List<String> =
        listOf("config", "set", setting.path) +
            if (setting.type == "array") text.split(Regex("[,\\s]+")).filter { it.isNotEmpty() } else listOf(text)

    /** `smith config unset <setting>`: back to the default. */
    fun unsetArguments(setting: SettingContract): List<String> = listOf("config", "unset", setting.path)

    /** What the person reads above the question: what the setting is for and why its default is what it is. */
    fun words(setting: SettingContract): String =
        listOf(
            setting.description,
            if (setting.reason.isEmpty()) "" else say(JETBRAINS_SETTINGS_WHY_THE_DEFAULT, setting.reason.toString()),
        ).filter(String::isNotEmpty).joinToString("\n")
}
