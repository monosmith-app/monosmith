package dev.monosmith.jetbrains.settings

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_ADDCHOICES_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.JETBRAINS_SETTINGS_WHY_THE_DEFAULT
import dev.monosmith.jetbrains.words.SmithWordsConfig.say

/**
 * One setting of smith.toml, as `smith config list --json` lists it.
 *
 * @property type `string`, `boolean`, `integer`, `array`...
 * @property reason why the default is what it is; empty when smith says none
 * @property value the value written in smith.toml, as JSON text; empty when the setting is not set
 * @property set true when smith.toml sets it
 * @property choices the values the setting accepts when it is an enum
 */
data class SettingContract(
    val path: String,
    val type: String,
    val description: String,
    val reason: String,
    val value: String,
    val set: Boolean,
    val choices: List<String>,
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
                    JsonFieldsMapper.words(it, "enum"),
                )
            }.filter { it.path.isNotEmpty() }

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

    /** The lines of a setting that takes one of a few values; the current one is first and marked. */
    fun enumRows(setting: SettingContract): List<ChoiceRowContract> {
        val current = shown(setting)
        return setting.choices.sortedByDescending { it == current }.map {
            ChoiceRowContract(
                it,
                it,
                if (it ==
                    current
                ) {
                    "(current)"
                } else {
                    ""
                },
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
