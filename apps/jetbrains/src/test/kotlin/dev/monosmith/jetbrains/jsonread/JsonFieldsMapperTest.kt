package dev.monosmith.jetbrains.jsonread

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNull
import kotlin.test.assertTrue

class JsonFieldsMapperTest {
    private val record =
        JsonFieldsMapper.parseObject(
            """{"a":"x","b":true,"n":3,"w":["p","q",1],"r":[{"k":"v"},2],"o":{"z":"y"},"nil":null}""",
        )

    @Test
    fun readsTextFlagsAndNumbers() {
        assertEquals("x", JsonFieldsMapper.text(record, "a"))
        assertTrue(JsonFieldsMapper.flag(record, "b"))
        assertEquals(3, JsonFieldsMapper.number(record, "n"))
    }

    @Test
    fun leavesAMissingOrMistypedFieldEmpty() {
        assertEquals("", JsonFieldsMapper.text(record, "missing"))
        assertEquals("", JsonFieldsMapper.text(record, "b"))
        assertEquals("", JsonFieldsMapper.text(record, "nil"))
        assertFalse(JsonFieldsMapper.flag(record, "a"))
        assertEquals(0, JsonFieldsMapper.number(record, "a"))
    }

    @Test
    fun readsListsKeepingOnlyTheRightType() {
        assertEquals(listOf("p", "q"), JsonFieldsMapper.words(record, "w"))
        assertEquals(1, JsonFieldsMapper.records(record, "r").size)
        assertEquals(emptyList(), JsonFieldsMapper.words(record, "a"))
    }

    @Test
    fun readsANestedObject() {
        assertEquals("y", JsonFieldsMapper.text(JsonFieldsMapper.record(record, "o")!!, "z"))
        assertNull(JsonFieldsMapper.record(record, "a"))
    }
}
