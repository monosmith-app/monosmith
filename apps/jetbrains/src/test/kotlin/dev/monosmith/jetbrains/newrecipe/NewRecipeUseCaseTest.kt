package dev.monosmith.jetbrains.newrecipe

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class NewRecipeUseCaseTest {
    private val fake = FakePortsHarness()
    private val leave = "\u0000leave"
    private val levels =
        """
        {"levels":[
          {"level":"language","additive":false,"plugins":[{"name":"node","description":"TS"},{"name":"go","description":"Go"}]},
          {"level":"lint","additive":true,"plugins":[{"name":"eslint","description":"e"},{"name":"golangci","description":"g"}]}],
         "fits":{"node":{"lint":["eslint"]},"go":{"lint":["golangci"]}}}
        """.trimIndent()

    init {
        fake.answers["recipe new --list-levels --json"] = fake.ok(levels)
    }

    @Test
    fun askesTheBasicsThenOneQuestionForEachLevelAndWritesOnlyAfterAYes() {
        fake.texts.addAll(listOf("team", "0.2.0", "base"))
        fake.ones.add("node")
        fake.manies.add(listOf("eslint"))
        fake.plans.add(fake.yes())
        assertTrue(NewRecipeUseCase.run(fake.ports))
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
                "lint=eslint",
            ),
            fake.calls.last(),
        )
        assertTrue(fake.shown.last())
        assertEquals("team.recipe.toml is written.", fake.infos.single())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun theLanguageQuestionStartsOnLeavingItToTheUsersAndTheLintListFollowsTheLanguage() {
        fake.texts.addAll(listOf("team", "0.1.0", ""))
        fake.ones.add("go")
        fake.manies.add(emptyList())
        fake.plans.add(fake.yes())
        NewRecipeUseCase.run(fake.ports)
        assertEquals(leave, fake.firstRowName())
        assertEquals(listOf("golangci"), fake.rowsSeen.last().map { it.name })
    }

    @Test
    fun theCommandIsShownBeforeAnythingIsWrittenAndCancelWritesNothing() {
        fake.texts.addAll(listOf("team", "0.1.0", ""))
        fake.ones.add(leave)
        fake.manies.add(emptyList())
        assertFalse(NewRecipeUseCase.run(fake.ports))
        assertEquals("smith recipe new --name team --recipe-version 0.1.0", fake.plansSeen.single())
        assertEquals(listOf(listOf("recipe", "new", "--list-levels", "--json")), fake.calls)
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun cancelAtTheNameWritesNothing() {
        assertFalse(NewRecipeUseCase.run(fake.ports))
        assertEquals(1, fake.calls.size)
    }

    @Test
    fun anOldSmithThatCannotListTheLevelsWritesNothing() {
        fake.answers["recipe new --list-levels --json"] = fake.fail("✖ unknown flag: --list-levels\n")
        assertFalse(NewRecipeUseCase.run(fake.ports))
        assertTrue(fake.warnings.single().startsWith("No recipe was written."))
    }

    @Test
    fun levelsItCannotReadAreAWarningThatSaysToUpdateThePlugin() {
        fake.answers["recipe new --list-levels --json"] = fake.ok("[1]")
        assertFalse(NewRecipeUseCase.run(fake.ports))
        assertTrue(fake.warnings.single().contains("Update the plugin"))
    }

    @Test
    fun aFailedWriteIsAWarningWithSmithsWords() {
        fake.texts.addAll(listOf("team", "0.1.0", ""))
        fake.ones.add("node")
        fake.manies.add(emptyList())
        fake.plans.add(fake.yes())
        fake.answers["recipe new --name team"] = fake.fail("✖ team.recipe.toml exists\n")
        assertFalse(NewRecipeUseCase.run(fake.ports))
        assertEquals("No recipe was written. team.recipe.toml exists", fake.warnings.single())
    }
}
