package dev.monosmith.jetbrains.trycommand

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class TryResultMapperTest {
    @Test
    fun aCommandAndEveryCheckThatPassedAreOfferedForReal() {
        val page =
            TryResultMapper.toPage(
                """{"command":["add","lib","x"],"commandOk":true,"changed":["a.txt"],"checks":[{"name":"build","ok":true}]}""",
            )
        assertTrue(page.passed)
        assertTrue(page.page.startsWith("Try: smith add lib x"))
        assertTrue(page.page.contains("  a.txt"))
        assertTrue(page.page.contains("build: passed"))
    }

    @Test
    fun aFailedCheckShowsItsTailAndIsNotOfferedForReal() {
        val page =
            TryResultMapper.toPage(
                """{"command":["add"],"commandOk":true,"checks":[{"name":"test","ok":false,"tail":"1 failed"}]}""",
            )
        assertFalse(page.passed)
        assertTrue(page.page.contains("test: failed"))
        assertTrue(page.page.contains("1 failed"))
        assertTrue(page.page.contains("Files it changed:\n  none"))
    }

    @Test
    fun aFailedCommandShowsItsTail() {
        val page = TryResultMapper.toPage("""{"command":["x"],"commandOk":false,"commandTail":"boom"}""")
        assertFalse(page.passed)
        assertTrue(page.page.contains("command: failed"))
        assertTrue(page.page.contains("boom"))
    }

    @Test
    fun somethingThatIsNotAResultIsRefused() {
        assertFailsWith<IllegalStateException> { TryResultMapper.toPage("""{"error":"x"}""") }
        assertEquals("Try:", TryResultMapper.toPage("""{"command":["a"]}""").page.take(4))
    }
}
