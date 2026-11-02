package dev.monosmith.jetbrains.agentfile

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class WriteAgentFileUseCaseTest {
    private val fake = FakePortsHarness()

    @Test
    fun theAgentFileIsPlannedWithDryRunThenWritten() {
        fake.answers["agent file --dry-run"] = fake.ok("would write AGENTS.md\n")
        fake.plans.add(fake.yes())
        assertTrue(WriteAgentFileUseCase.run(fake.ports))
        assertEquals("would write AGENTS.md", fake.plansSeen.single())
        assertEquals(listOf("agent", "file", "--yes"), fake.calls.last())
    }

    @Test
    fun cancelWritesNothing() {
        assertFalse(WriteAgentFileUseCase.run(fake.ports))
        assertTrue(fake.calls.none { "--yes" in it })
    }
}
