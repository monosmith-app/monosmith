package dev.monosmith.jetbrains.newrecipe

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class RecipeLevelsMapperTest {
    private val json =
        """
        {"levels":[
          {"level":"language","additive":false,"plugins":[{"name":"node","description":"TS"},{"name":"go","description":"Go"}]},
          {"level":"lint","additive":true,"plugins":[{"name":"eslint","description":"e"},{"name":"golangci","description":"g"}]},
          {"level":"","plugins":[]}],
         "fits":{"node":{"lint":["eslint"]},"go":{"lint":["golangci"]}}}
        """.trimIndent()

    @Test
    fun readsTheLevelsAndWhatFitsEachLanguage() {
        val levels = RecipeLevelsMapper.toLevels(json)
        assertEquals(listOf("language", "lint"), levels.levels.map { it.level })
        assertTrue(levels.levels.last().additive)
        assertEquals(listOf("eslint"), levels.fits.getValue("node").getValue("lint"))
    }

    @Test
    fun afterALanguageOnlyItsPluginsAreOffered() {
        val levels = RecipeLevelsMapper.toLevels(json)
        val lint = levels.levels.last()
        assertEquals(listOf("eslint", "golangci"), RecipeLevelsMapper.optionsFor(levels, lint, "").map { it.name })
        assertEquals(listOf("golangci"), RecipeLevelsMapper.optionsFor(levels, lint, "go").map { it.name })
        assertEquals(listOf("node", "go"), RecipeLevelsMapper.optionsFor(levels, levels.levels.first(), "go").map { it.name })
        assertEquals(emptyList(), RecipeLevelsMapper.optionsFor(levels, lint, "rust"))
    }

    @Test
    fun givesOneFixedFlagForEveryLevelThatHasPlugins() {
        val request =
            RecipeRequestContract(
                "team",
                "0.2.0",
                "base",
                linkedMapOf(
                    "language" to listOf("node"),
                    "lint" to listOf("eslint", "prettier"),
                    "test" to emptyList(),
                ),
            )
        assertEquals(
            listOf(
                "recipe",
                "new",
                "--name",
                "team",
                "--recipe-version",
                "0.2.0",
                "--extends",
                "base",
                "--fixed",
                "language=node",
                "--fixed",
                "lint=eslint,prettier",
            ),
            RecipeLevelsMapper.arguments(request),
        )
        assertEquals(listOf("recipe", "new", "--name", "t"), RecipeLevelsMapper.arguments(RecipeRequestContract("t", "", "", emptyMap())))
        assertEquals("team.recipe.toml", RecipeLevelsMapper.file(request))
    }

    @Test
    fun checksTheNameAndTheVersion() {
        assertNull(RecipeLevelsMapper.nameProblem("team-node"))
        assertNotNull(RecipeLevelsMapper.nameProblem("Team"))
        assertNull(RecipeLevelsMapper.versionProblem(""))
        assertNull(RecipeLevelsMapper.versionProblem("0.1.0"))
        assertNotNull(RecipeLevelsMapper.versionProblem("1.0"))
    }
}
