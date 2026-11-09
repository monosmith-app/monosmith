package dev.monosmith.jetbrains.smithrun

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class FailureTextMapperTest {
    @Test
    fun usesStandardErrorWithItsFix() {
        val run = SmithRunContract(1, "", "× the tree is not clean\n→ commit or stash first\n")
        assertEquals("the tree is not clean Try: commit or stash first", FailureTextMapper.describe(run))
    }

    @Test
    fun fallsBackToStandardOutputThenToTheExitCode() {
        assertEquals("nothing to do", FailureTextMapper.describe(SmithRunContract(1, "nothing to do\n", "")))
        assertEquals("smith stopped with exit code 3", FailureTextMapper.describe(SmithRunContract(3, "", "")))
    }

    @Test
    fun theRunKnowsItsLinesAndItsSuccess() {
        val run = SmithRunContract(0, "a\n\n  b  \n", "")
        assertTrue(run.ok)
        assertEquals(listOf("a", "b"), run.lines)
    }
}
