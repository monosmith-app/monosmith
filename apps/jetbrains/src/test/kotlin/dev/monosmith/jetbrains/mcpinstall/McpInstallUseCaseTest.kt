package dev.monosmith.jetbrains.mcpinstall

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class McpInstallUseCaseTest {
    private val fake = FakePortsHarness()

    @Test
    fun theAssistantsStartUntickedAndOnlyTheTickedOnesAreWritten() {
        fake.manies.add(listOf("--claude"))
        fake.plans.add(fake.yes())
        assertTrue(McpInstallUseCase.run(fake.ports))
        assertEquals(listOf(false, false), McpInstallUseCase.rows.map { it.picked })
        assertEquals(listOf("mcp", "install", "--claude", "--yes"), fake.calls.last())
        assertTrue(fake.plansSeen.single().startsWith("smith mcp install --claude"))
    }

    @Test
    fun tickingNoAssistantWritesNothing() {
        fake.manies.add(emptyList())
        assertFalse(McpInstallUseCase.run(fake.ports))
        assertEquals(emptyList(), fake.calls)
    }

    @Test
    fun cancelAtThePlanWritesNothing() {
        fake.manies.add(listOf("--claude", "--vscode"))
        assertFalse(McpInstallUseCase.run(fake.ports))
        assertTrue(fake.calls.none { "--yes" in it })
    }
}
