package dev.monosmith.jetbrains.newworkspace

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class NewQuestionsMapperTest {
    private val json =
        """
        [{"key":"ci","prompt":"CI?","default":"github","flag":"ci","many":true,"options":[{"value":"github","label":"GitHub Actions","hint":"default"},{"value":"azure","label":"Azure"}]},
         {"key":"node-stack","prompt":"Node?","default":"node-jest","when":{"language":"node"},"options":[{"value":"node-jest","label":"Jest","hint":"TypeScript and Jest (default)"}]},
         {"key":"release-branch","prompt":"Branch?","default":"main","advanced":true},
         {"key":"save-recipe","prompt":"Save?","default":"no","atEnd":true,"options":[{"value":"no","label":"No"}]},
         {"key":"","prompt":"x"}]
        """.trimIndent()
    private val questions = NewQuestionsMapper.toQuestions(json)

    @Test
    fun readsTheQuestionsAndDropsOneWithNoKey() {
        assertEquals(listOf("ci", "node-stack", "release-branch", "save-recipe"), questions.map { it.key })
        assertTrue(questions.first().many)
        assertEquals(mapOf("language" to "node"), questions[1].`when`)
    }

    @Test
    fun theDefaultIsMarkedAndTheHintLosesItsDefaultWord() {
        val rows = NewQuestionsMapper.rows(questions.first())
        assertEquals(listOf("(default)", ""), rows.map { it.description })
        assertEquals(listOf(true, false), rows.map { it.picked })
        assertEquals("", rows.first().detail)
        assertEquals("TypeScript and Jest", NewQuestionsMapper.rows(questions[1]).single().detail)
    }

    @Test
    fun theMainQuestionsLeaveOutTheAdvancedTheFreeTextAndTheOnesAskedAtTheEnd() {
        assertEquals(listOf("ci", "node-stack"), NewQuestionsMapper.main(questions).map { it.key })
    }

    @Test
    fun aQuestionAppliesWhenItsAnswersAreGiven() {
        assertTrue(NewQuestionsMapper.applies(questions[1], mapOf("language" to "node")))
        assertFalse(NewQuestionsMapper.applies(questions[1], mapOf("language" to "go")))
        assertTrue(NewQuestionsMapper.applies(questions[0], emptyMap()))
    }

    @Test
    fun labelsAnAnswerOrKeepsIt() {
        assertEquals("Azure", NewQuestionsMapper.labelOf(questions.first(), "azure"))
        assertEquals("other", NewQuestionsMapper.labelOf(questions.first(), "other"))
        assertEquals("x", NewQuestionsMapper.labelOf(null, "x"))
    }
}
