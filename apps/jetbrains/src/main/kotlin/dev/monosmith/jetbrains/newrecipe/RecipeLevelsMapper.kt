package dev.monosmith.jetbrains.newrecipe

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper
import dev.monosmith.jetbrains.prompt.ChoiceRowContract

/** A plugin a recipe can fix at a level. */
data class RecipePluginContract(
    val name: String,
    val description: String,
)

/**
 * One level of a recipe.
 *
 * @property additive the level takes several plugins (a checkbox list); false takes one (radio buttons)
 */
data class RecipeLevelContract(
    val level: String,
    val additive: Boolean,
    val plugins: List<RecipePluginContract>,
)

/**
 * What `smith recipe new --list-levels --json` lists.
 *
 * @property fits for each language, the plugin names that fit each level
 */
data class RecipeLevelsContract(
    val levels: List<RecipeLevelContract>,
    val fits: Map<String, Map<String, List<String>>>,
)

/** What the person decided about a recipe: the basics and the plugin fixed at each level. */
data class RecipeRequestContract(
    val name: String,
    val version: String,
    val extends: String,
    val fixed: Map<String, List<String>>,
)

/** The reader of the levels of a recipe, the questions about them and the command line of `smith recipe new`. */
object RecipeLevelsMapper {
    private val namePattern = Regex("^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$")
    private val versionPattern = Regex("^\\d+\\.\\d+\\.\\d+$")

    /**
     * @param json what `smith recipe new --list-levels --json` printed
     * @return the levels that have a name, and the plugins that fit each language
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun toLevels(json: String): RecipeLevelsContract {
        val root = JsonFieldsMapper.parseObject(json)
        val levels =
            JsonFieldsMapper
                .records(root, "levels")
                .map { level ->
                    RecipeLevelContract(
                        JsonFieldsMapper.text(level, "level"),
                        JsonFieldsMapper.flag(level, "additive"),
                        JsonFieldsMapper
                            .records(level, "plugins")
                            .map { RecipePluginContract(JsonFieldsMapper.text(it, "name"), JsonFieldsMapper.text(it, "description")) }
                            .filter { it.name.isNotEmpty() },
                    )
                }.filter { it.level.isNotEmpty() }
        val fits =
            JsonFieldsMapper.record(root, "fits")?.entrySet().orEmpty().associate { (language, byLevel) ->
                language to
                    (
                        byLevel
                            .takeIf { it.isJsonObject }
                            ?.asJsonObject
                            ?.entrySet()
                            .orEmpty()
                            .associate { (level, names) ->
                                level to
                                    (
                                        names
                                            .takeIf { it.isJsonArray }
                                            ?.asJsonArray
                                            ?.filter { it.isJsonPrimitive }
                                            ?.map { it.asString }
                                            .orEmpty()
                                    )
                            }
                    )
            }
        return RecipeLevelsContract(levels, fits)
    }

    /**
     * The plugins to offer at a level, or none when no plugin can fill it. After a language is fixed only
     * the plugins that fit it are offered.
     *
     * @param language the language plugin fixed so far, or empty
     */
    fun optionsFor(
        levels: RecipeLevelsContract,
        level: RecipeLevelContract,
        language: String,
    ): List<RecipePluginContract> {
        if (language.isEmpty() || level.level == "language") return level.plugins
        val names = levels.fits[language]?.get(level.level).orEmpty()
        return level.plugins.filter { it.name in names }
    }

    /** The lines of the question about a level. A level of one answer starts on "Leave it to the users of the recipe". */
    fun rows(options: List<RecipePluginContract>): List<ChoiceRowContract> =
        options.map { ChoiceRowContract(it.name, it.name, "", it.description) }

    /** The arguments of `smith recipe new`: the name, the version and the parent when given, and one `--fixed level=plugin` for each level. */
    fun arguments(request: RecipeRequestContract): List<String> =
        buildList {
            addAll(listOf("recipe", "new", "--name", request.name))
            if (request.version.isNotEmpty()) addAll(listOf("--recipe-version", request.version))
            if (request.extends.isNotEmpty()) addAll(listOf("--extends", request.extends))
            request.fixed.forEach { (level, plugins) ->
                if (plugins.isNotEmpty()) addAll(listOf("--fixed", "$level=${plugins.joinToString(",")}"))
            }
        }

    /** The file `smith recipe new` writes: it is named after the recipe, in the workspace folder. */
    fun file(request: RecipeRequestContract): String = "${request.name}.recipe.toml"

    /** What is wrong with a recipe name, or null when it is fine. */
    fun nameProblem(name: String): String? =
        if (namePattern.matches(
                name,
            )
        ) {
            null
        } else {
            "Use lowercase words joined by dashes, for example \"team\" or \"team-node\". The name is the name of the recipe's releases."
        }

    /** What is wrong with a version, or null when it is fine. Empty takes 0.1.0. */
    fun versionProblem(version: String): String? =
        if (version.isEmpty() ||
            versionPattern.matches(version)
        ) {
            null
        } else {
            "Use major.minor.patch, for example 0.1.0."
        }
}
