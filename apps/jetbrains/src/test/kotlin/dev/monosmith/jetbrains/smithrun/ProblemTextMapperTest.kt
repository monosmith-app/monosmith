package dev.monosmith.jetbrains.workspace

import kotlin.test.Test
import kotlin.test.assertEquals

class ProblemTextMapperTest {
    @Test
    fun keepsTheProblemAndTheFirstRemedy() {
        val stderr = "✖ smith.toml:3:9: expected a value\n  → fix the line, then run smith doctor\n  → or restore it with git\n"
        assertEquals("smith.toml:3:9: expected a value Try: fix the line, then run smith doctor", ProblemTextMapper.toLine(stderr))
    }

    @Test
    fun keepsAProblemWithoutARemedy() {
        assertEquals("smith exited", ProblemTextMapper.toLine("\nsmith exited\n"))
    }

    @Test
    fun givesNothingForEmptyOutput() {
        assertEquals("", ProblemTextMapper.toLine("  \n"))
    }
}
