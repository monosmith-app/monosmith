package dev.monosmith.jetbrains.pluginswitch

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/**
 * One built-in plugin of smith.
 *
 * @property name the name `smith plugin enable` and `disable` take
 * @property roles what the plugin does (language, toolchain, scaffold)
 */
data class BuiltinPluginContract(
    val name: String,
    val roles: List<String>,
)

/** Reads `smith plugin list --json` and words the built-in plugins as the rows of a question. */
object BuiltinPluginsMapper {
    /**
     * @param json what smith printed
     * @return the built-in plugins, in smith's order
     * @throws IllegalStateException when the text is not a JSON list
     */
    fun toPlugins(json: String): List<BuiltinPluginContract> =
        JsonFieldsMapper
            .parseArray(json)
            .filter { it.isJsonObject }
            .map { it.asJsonObject }
            .filter { JsonFieldsMapper.text(it, "source") == "built-in" }
            .map { BuiltinPluginContract(JsonFieldsMapper.text(it, "name"), JsonFieldsMapper.words(it, "roles")) }
            .filter { it.name.isNotEmpty() }

    /** One row for each plugin; its roles are the detail. */
    fun rows(plugins: List<BuiltinPluginContract>): List<ChoiceRowContract> =
        plugins.map { ChoiceRowContract(it.name, it.name, "", it.roles.joinToString(", ")) }
}
