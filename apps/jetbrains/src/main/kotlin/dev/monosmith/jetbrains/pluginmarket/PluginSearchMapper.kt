package dev.monosmith.jetbrains.pluginmarket

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/**
 * A plugin of the marketplace, as `smith plugin search --json` lists it.
 *
 * @property newest its newest stable version
 * @property roles what it does in a project (`scaffold`, `deploy`)
 */
data class PluginContract(
    val name: String,
    val newest: String,
    val description: String,
    val roles: List<String>,
    val languages: List<String>,
    val license: String,
    val homepage: String,
)

/** The command lines and the lists of searching the marketplace and installing a plugin. */
object PluginSearchMapper {
    /**
     * @param json what smith printed
     * @return the plugins that have a name
     * @throws IllegalStateException when the text is not a JSON list
     */
    fun toPlugins(json: String): List<PluginContract> =
        JsonFieldsMapper
            .parseArray(json)
            .filter { it.isJsonObject }
            .map { it.asJsonObject }
            .map {
                PluginContract(
                    JsonFieldsMapper.text(it, "name"),
                    JsonFieldsMapper.text(it, "newest"),
                    JsonFieldsMapper.text(it, "description"),
                    JsonFieldsMapper.words(it, "roles"),
                    JsonFieldsMapper.words(it, "languages"),
                    JsonFieldsMapper.text(it, "license"),
                    JsonFieldsMapper.text(it, "homepage"),
                )
            }.filter { it.name.isNotEmpty() }

    /** `smith plugin search [words] --json`; the words are separate arguments. */
    fun searchArguments(term: String): List<String> =
        listOf("plugin", "search") + term.split(Regex("\\s+")).filter { it.isNotEmpty() } + "--json"

    /** The lines of "which plugin?". */
    fun rows(plugins: List<PluginContract>): List<ChoiceRowContract> =
        plugins.map {
            ChoiceRowContract(
                it.name,
                "${it.name} ${it.newest}",
                listOf(it.languages.joinToString(", "), it.roles.joinToString(", ")).filter(String::isNotEmpty).joinToString(" · "),
                it.description,
            )
        }

    /**
     * `smith plugin install <name>@<version>`. The version is the one the person saw. [trust] records that
     * the person trusts the plugin in this workspace; [pin] pins the sha256 of the program it runs. Both are
     * off unless asked.
     */
    fun installArguments(
        plugin: PluginContract,
        trust: Boolean,
        pin: Boolean,
    ): List<String> =
        listOf("plugin", "install", if (plugin.newest.isEmpty()) plugin.name else "${plugin.name}@${plugin.newest}") +
            (if (trust) listOf("--trust") else emptyList()) +
            (if (pin) listOf("--pin") else emptyList())

    private fun license(plugin: PluginContract): String = if (plugin.license.isEmpty()) "" else " (license ${plugin.license})"

    /** What the install checks and what the two boxes mean, said before the question. */
    fun trustWords(plugin: PluginContract): String =
        "${plugin.name} ${plugin.newest} is a program that runs on your computer${license(plugin)}. " +
            "smith checks the archive's SHA-256 and file list and every path before it installs. " +
            "Installing does not trust it: --trust records that you trust it in this workspace. Nothing is committed."
}
