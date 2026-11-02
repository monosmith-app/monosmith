package dev.monosmith.jetbrains.addchoices

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ChoiceRowsAlgorithmTest {
    private fun choice(
        name: String,
        axis: String,
        language: String = "node",
        isDefault: Boolean = false,
        many: Boolean = false,
        needs: List<String> = emptyList(),
        suggested: Boolean = false,
        conflicts: List<String> = emptyList(),
        layer: String = "",
        requires: List<String> = emptyList(),
    ) = ChoiceOptionContract(axis, name, "the $name", language, requires, isDefault, many, needs, suggested, conflicts, "", layer)

    private val catalogue =
        listOf(
            choice("react", "framework"),
            choice("vue", "framework"),
            choice("jest", "test", isDefault = true),
            choice("vitest", "test"),
            choice("electron", "shell"),
            choice("tauri", "shell"),
            choice("azure-functions", "framework", needs = listOf("trigger", "runtime"), layer = "host"),
            choice("http", "trigger", suggested = true),
            choice("timer", "trigger"),
            choice("node20", "runtime", conflicts = listOf("timer")),
            choice("gofmt", "lint", language = "go"),
            choice("chakra", "library", many = true, requires = listOf("react")),
        )

    @Test
    fun groupsTheChoicesOfALanguageByAxisInSmithsOrder() {
        val axes = ChoiceRowsAlgorithm.byAxis(catalogue, "node")
        assertEquals(listOf("framework", "test", "shell", "trigger", "runtime", "library"), axes.map { it.axis })
        assertTrue(axes.first { it.axis == "library" }.many)
        assertFalse(axes.first { it.axis == "test" }.many)
    }

    @Test
    fun leavesOutTheChoicesOfAnotherLanguage() {
        assertEquals(listOf("lint"), ChoiceRowsAlgorithm.byAxis(catalogue, "go").map { it.axis })
    }

    @Test
    fun anAxisWithNoDefaultThatAChoiceNeedsIsAsked() {
        assertEquals(listOf("trigger", "runtime"), ChoiceRowsAlgorithm.askedAxes(catalogue, "node"))
    }

    @Test
    fun aHostFitsOnlyAFunctionAppAndAWrappingShellOnlyADesktopApp() {
        val host = choice("azure-functions", "framework", layer = "host")
        assertTrue(ChoiceRowsAlgorithm.fitsOutcome(host, "function"))
        assertTrue(ChoiceRowsAlgorithm.fitsOutcome(host, null))
        assertFalse(ChoiceRowsAlgorithm.fitsOutcome(host, "web"))
        val wrapper = choice("electron", "framework", layer = "shell")
        assertTrue(ChoiceRowsAlgorithm.fitsOutcome(wrapper, "desktop"))
        assertFalse(ChoiceRowsAlgorithm.fitsOutcome(wrapper, "web"))
    }

    @Test
    fun aSingleChoiceListStartsOnItsDefault() {
        val test = ChoiceRowsAlgorithm.byAxis(catalogue.reversed(), "node").first { it.axis == "test" }
        val rows = ChoiceRowsAlgorithm.singleChoiceRows(test, emptyList(), null)
        assertEquals(listOf("jest", "vitest"), rows.map { it.name })
        assertEquals("(default)", rows.first().description)
    }

    @Test
    fun anAxisWithNoDefaultStartsOnNoneSoEnterAddsNothing() {
        val shell = ChoiceRowsAlgorithm.byAxis(catalogue, "node").first { it.axis == "shell" }
        val rows = ChoiceRowsAlgorithm.singleChoiceRows(shell, emptyList(), null, mapOf("shell" to "Desktop shell"))
        assertEquals(listOf("", "electron", "tauri"), rows.map { it.name })
        assertTrue(rows.first().detail.contains("desktop shell"))
    }

    @Test
    fun aChoiceThatConflictsWithWhatIsPickedIsNotListed() {
        val runtime = ChoiceRowsAlgorithm.byAxis(catalogue, "node").first { it.axis == "runtime" }
        val timer = catalogue.first { it.name == "timer" }
        assertEquals(emptyList(), ChoiceRowsAlgorithm.singleChoiceRows(runtime, listOf(timer), null).filter { it.name == "node20" })
    }

    @Test
    fun theSuggestedAnswerOfAnAskedAxisComesFirst() {
        val trigger = catalogue.filter { it.axis == "trigger" }.reversed()
        val rows = ChoiceRowsAlgorithm.askedRows(trigger, emptyList())
        assertEquals(listOf("http", "timer"), rows.map { it.name })
        assertEquals("(suggested)", rows.first().description)
    }

    @Test
    fun theDefaultsOfAManyAxisStartTicked() {
        val axis =
            AxisGroupContract(
                "lint",
                true,
                listOf(choice("eslint", "lint", isDefault = true, many = true), choice("prettier", "lint", many = true)),
            )
        assertEquals(listOf(true, false), ChoiceRowsAlgorithm.manyChoiceRows(axis, emptyList()).map { it.picked })
    }

    @Test
    fun saysTheChoicesInWordsWithTheDefaultsMarked() {
        val words = ChoiceRowsAlgorithm.inWords(listOf("react", "jest"), null, "node", catalogue, mapOf("shell" to "Desktop shell"))
        assertEquals("Your choices: react, jest (default); no desktop shell.", words)
    }

    @Test
    fun saysNothingExtraWasChosenAndNamesAStack() {
        assertEquals(
            "Built with the default choices of node: nothing extra was chosen.",
            ChoiceRowsAlgorithm.inWords(emptyList(), null, "node", catalogue),
        )
        assertEquals("Built from the stack web.", ChoiceRowsAlgorithm.inWords(listOf("react"), "web", "node", catalogue))
    }
}
