package dev.monosmith.jetbrains.readcommands

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ReadSmithCommandUseCaseTest {
    private val fake = FakePortsHarness()

    @Test
    fun aReadingCommandPrintsInTheOutputTabAndAsksNothing() {
        assertTrue(ReadSmithCommandUseCase.run(fake.ports, "the cache status", listOf("cache", "status")))
        assertEquals(listOf(listOf("cache", "status")), fake.calls)
        assertEquals(listOf(true), fake.shown)
        assertEquals(emptyList(), fake.plansSeen)
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aFailedReadingIsAWarningWithSmithsWords() {
        fake.answers["cache status"] = fake.fail("× the cache folder cannot be read\n")
        assertFalse(ReadSmithCommandUseCase.run(fake.ports, "the cache status", listOf("cache", "status")))
        assertEquals("smith could not show the cache status. the cache folder cannot be read", fake.warnings.single())
    }

    @Test
    fun whyMissTakesProjectColonTaskAndPassesItAsOneArgument() {
        fake.texts.add("web:build")
        assertTrue(ReadSmithCommandUseCase.whyMiss(fake.ports))
        assertEquals(listOf("why-miss", "web:build"), fake.calls.single())
    }

    @Test
    fun cancelAtTheTaskRunsNothing() {
        assertFalse(ReadSmithCommandUseCase.whyMiss(fake.ports))
        assertEquals(emptyList(), fake.calls)
    }
}
