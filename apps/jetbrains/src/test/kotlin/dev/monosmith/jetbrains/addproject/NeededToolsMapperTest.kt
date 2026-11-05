package dev.monosmith.jetbrains.addproject

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull

class NeededToolsMapperTest {
    @Test
    fun readsTheToolsAndTheProblemOfEach() {
        val output =
            """
            Wrote 12 file(s):
            ! golangci-lint is 2.13.0, not the 2.14.0 that CI pins: smith lint of the go projects fails.
            ! Python is not installed: smith build fails.
            → Run smith tools install golangci-lint python to choose how to get them.
            """.trimIndent()
        val tools = NeededToolsMapper.parse(output)!!
        assertEquals(listOf("golangci-lint", "python"), tools.names)
        assertEquals(
            listOf(
                "golangci-lint is 2.13.0, not the 2.14.0 that CI pins: smith lint of the go projects fails.",
                "Python is not installed: smith build fails.",
            ),
            tools.problems,
        )
    }

    @Test
    fun saysNothingWhenNoToolIsLacking() {
        assertNull(NeededToolsMapper.parse("Wrote 3 file(s):\nDone: app web is ready."))
    }
}
