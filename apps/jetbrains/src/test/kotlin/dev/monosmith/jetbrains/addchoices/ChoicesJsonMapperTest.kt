package dev.monosmith.jetbrains.addchoices

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ChoicesJsonMapperTest {
    @Test
    fun readsAChoiceWithEveryField() {
        val json =
            """[{"axis":"framework","name":"aws-lambda","description":"a function","title":"AWS Lambda",""" +
                """"layer":"host","language":"dotnet",""" +
                """"needs":["trigger","runtime"],"conflicts":["x"],"requires":["y"],"default":true,"many":true,"suggested":true}]"""
        val choice = ChoicesJsonMapper.toChoices(json).single()
        assertEquals("AWS Lambda", choice.label)
        assertEquals(listOf("trigger", "runtime"), choice.needs)
        assertEquals(listOf("x"), choice.conflicts)
        assertEquals(listOf("y"), choice.requires)
        assertTrue(choice.isDefault && choice.many && choice.suggested)
        assertEquals("host", choice.layer)
    }

    @Test
    fun theNameStandsInWhenThereIsNoTitle() {
        assertEquals("vue", ChoicesJsonMapper.toChoices("""[{"axis":"framework","name":"vue"}]""").single().label)
    }

    @Test
    fun dropsAnEntryWithNoAxisOrNoName() {
        assertEquals(emptyList(), ChoicesJsonMapper.toChoices("""[{"name":"x"},{"axis":"a"},"text"]"""))
    }
}
