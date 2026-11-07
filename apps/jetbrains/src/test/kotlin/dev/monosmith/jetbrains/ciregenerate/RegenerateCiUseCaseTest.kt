package dev.monosmith.jetbrains.ciregenerate

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class RegenerateCiUseCaseTest {
    private val fake = FakePortsHarness()

    @Test
    fun showsTheDiffOfEveryFileBeforeThePipelinesAreRegenerated() {
        fake.answers["ci generate --dry-run"] = fake.ok("- .github/workflows/old.yml\n+ .github/workflows/ci.yml\n")
        fake.plans.add(fake.yes())
        assertTrue(RegenerateCiUseCase.run(fake.ports))
        assertTrue(fake.plansSeen.single().contains("+ .github/workflows/ci.yml"))
        assertEquals(listOf("ci", "generate", "--yes"), fake.calls.last())
    }

    @Test
    fun cancelWritesNothing() {
        assertFalse(RegenerateCiUseCase.run(fake.ports))
        assertTrue(fake.calls.none { "--yes" in it })
    }

    @Test
    fun aRefusedPlanIsAWarning() {
        fake.answers["ci generate --dry-run"] = fake.fail("× smith.toml has no ci table\n")
        assertFalse(RegenerateCiUseCase.run(fake.ports))
        assertEquals("Nothing was changed. smith.toml has no ci table", fake.warnings.single())
    }
}
