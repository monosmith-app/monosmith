package dev.monosmith.jetbrains.addchoices

import kotlin.test.Test
import kotlin.test.assertEquals

class GlossaryAlgorithmTest {
    private val titles = mapOf("shell" to "Desktop shell")
    private val about =
        mapOf(
            "shell" to "what wraps the app as a program you install on a computer.",
            "test" to "the tool that runs the tests",
        )

    @Test
    fun saysWhatTheLayerMeansInTheCataloguesWords() {
        assertEquals(
            "desktop shell = what wraps the app as a program you install on a computer.",
            GlossaryAlgorithm.line("shell", titles, about),
        )
    }

    @Test
    fun theIdStandsInForAMissingTitleAndTheSentenceEndsOnce() {
        assertEquals("test = the tool that runs the tests.", GlossaryAlgorithm.line("test", titles, about))
    }

    @Test
    fun saysNothingWhenTheCatalogueSaysNothing() {
        assertEquals("", GlossaryAlgorithm.line("style", titles, about))
        assertEquals("Pick one.", GlossaryAlgorithm.withLine("Pick one.", ""))
        assertEquals("Pick one.\nx = y.", GlossaryAlgorithm.withLine("Pick one.", "x = y."))
    }
}
