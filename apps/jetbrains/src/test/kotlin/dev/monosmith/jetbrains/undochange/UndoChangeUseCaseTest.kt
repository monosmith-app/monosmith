package dev.monosmith.jetbrains.undochange

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class UndoChangeUseCaseTest {
    private val fake = FakePortsHarness()

    init {
        fake.answers["undo --dry-run"] = fake.ok("would revert abc123 add app web\n")
    }

    @Test
    fun showsWhatItWouldRevertAndRevertsOnlyAfterAYes() {
        fake.plans.add(fake.yes())
        fake.answers["undo --yes"] = fake.ok("reverted abc123\n")
        assertTrue(UndoChangeUseCase.run(fake.ports))
        assertEquals("would revert abc123 add app web", fake.plansSeen.single())
        assertEquals(listOf("undo", "--yes"), fake.calls.last())
        assertEquals("reverted abc123", fake.infos.single())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun cancelRevertsNothing() {
        assertFalse(UndoChangeUseCase.run(fake.ports))
        assertEquals(listOf(listOf("undo", "--dry-run")), fake.calls)
    }

    @Test
    fun smithRefusesWhenALaterChangeTouchesTheSameFilesAndTheWarningSaysWhy() {
        fake.answers["undo --dry-run"] = fake.fail("✖ a later commit changes the same files\n→ undo that one first\n")
        assertFalse(UndoChangeUseCase.run(fake.ports))
        assertEquals("Nothing was undone. a later commit changes the same files Try: undo that one first", fake.warnings.single())
        assertEquals(emptyList(), fake.plansSeen)
    }

    @Test
    fun aFailedRevertIsAWarning() {
        fake.plans.add(fake.yes())
        fake.answers["undo --yes"] = fake.fail("✖ the revert has a conflict\n")
        assertFalse(UndoChangeUseCase.run(fake.ports))
        assertTrue(fake.warnings.single().startsWith("Nothing was undone."))
    }
}
