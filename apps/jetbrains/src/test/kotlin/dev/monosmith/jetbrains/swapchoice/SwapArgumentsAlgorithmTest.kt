package dev.monosmith.jetbrains.swapchoice

import dev.monosmith.jetbrains.addchoices.AxisGroupContract
import dev.monosmith.jetbrains.addchoices.ChoiceOptionContract
import kotlin.test.Test
import kotlin.test.assertEquals

class SwapArgumentsAlgorithmTest {
    private fun choice(
        name: String,
        isDefault: Boolean = false,
    ) = ChoiceOptionContract("test", name, "the $name", "node", emptyList(), isDefault, false, emptyList(), false, emptyList(), "", "")

    @Test
    fun thePlanIsADryRunAndTheChangeAnswersYes() {
        assertEquals(
            listOf("swap", "web", "test", "vitest", "--dry-run"),
            SwapArgumentsAlgorithm.arguments("web", "test", "vitest", apply = false),
        )
        assertEquals(
            listOf("swap", "web", "test", "vitest", "--yes"),
            SwapArgumentsAlgorithm.arguments("web", "test", "vitest", apply = true),
        )
    }

    @Test
    fun theAxesAndTheAnswersComeFromTheCatalogue() {
        val axis = AxisGroupContract("test", false, listOf(choice("jest", isDefault = true), choice("vitest")))
        val axes = SwapArgumentsAlgorithm.axisRows(listOf(axis), mapOf("test" to "Test runner"))
        assertEquals("Test runner", axes.single().label)
        assertEquals("2 answers", axes.single().description)
        val answers = SwapArgumentsAlgorithm.choiceRows(axis)
        assertEquals(listOf("(default)", ""), answers.map { it.description })
    }
}
