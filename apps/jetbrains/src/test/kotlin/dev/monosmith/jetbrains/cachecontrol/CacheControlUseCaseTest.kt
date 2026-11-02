package dev.monosmith.jetbrains.cachecontrol

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class CacheControlUseCaseTest {
    private val fake = FakePortsHarness()

    @Test
    fun trimmingShowsThePlanOfDryRunAndThenTrims() {
        fake.answers["cache prune --dry-run"] = fake.ok("would remove 12 entries; 640 MB remain\n")
        fake.plans.add(fake.yes())
        assertTrue(CacheControlUseCase.run(fake.ports, CacheControlUseCase.Kind.PRUNE))
        assertEquals("would remove 12 entries; 640 MB remain", fake.plansSeen.single())
        assertEquals(listOf("cache", "prune"), fake.calls.last())
        assertEquals("The cache is trimmed. Nothing is committed.", fake.infos.single())
    }

    @Test
    fun deletingTheCacheShowsWhatItCostsAndAsksBeforeItRuns() {
        fake.plans.add(fake.yes())
        assertTrue(CacheControlUseCase.run(fake.ports, CacheControlUseCase.Kind.CLEAR))
        assertTrue(fake.plansSeen.single().startsWith("smith cache clear"))
        assertTrue(fake.plansSeen.single().contains("The remote cache stays"))
        assertEquals(listOf("cache", "clear", "--yes"), fake.calls.last())
    }

    @Test
    fun removingWhatIsBrokenRunsVerifyWithFixOnlyAfterAYes() {
        fake.plans.add(fake.yes())
        assertTrue(CacheControlUseCase.run(fake.ports, CacheControlUseCase.Kind.REPAIR))
        assertEquals(listOf("cache", "verify", "--fix"), fake.calls.last())
        assertTrue(fake.plansSeen.single().contains("corrupt blobs"))
    }

    @Test
    fun cancelChangesNothingForEveryKind() {
        for (kind in CacheControlUseCase.Kind.entries) {
            assertFalse(CacheControlUseCase.run(fake.ports, kind))
        }
        assertTrue(fake.calls.none { "--yes" in it || it == listOf("cache", "prune") || it == listOf("cache", "verify", "--fix") })
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aFailedChangeIsAWarningWithSmithsWords() {
        fake.plans.add(fake.yes())
        fake.answers["cache clear --yes"] = fake.fail("✖ the cache is in use\n")
        assertFalse(CacheControlUseCase.run(fake.ports, CacheControlUseCase.Kind.CLEAR))
        assertEquals("Nothing was changed. the cache is in use", fake.warnings.single())
    }
}
