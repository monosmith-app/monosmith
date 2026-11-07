package dev.monosmith.jetbrains.derivefiles

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class DeriveFilesUseCaseTest {
    private val fake = FakePortsHarness()
    private val all = "\u0000all"
    private var committed = ""
    private val check =
        """
        2 derived files need action:
          Drift      packages/web/tsconfig.json  (tsconfig-references)  edited by hand
          Stale      smith.lock  (lock)  inputs changed
        5 derived files in all
        """.trimIndent()

    private fun derive() = DeriveFilesUseCase.run(fake.ports) { committed = it }

    init {
        fake.answers["derive --check"] = fake.ok(check)
    }

    @Test
    fun withNothingToDoItSaysEveryFileIsCurrent() {
        fake.answers["derive --check"] = fake.ok("")
        assertFalse(derive())
        assertEquals("Every derived file is current.", fake.infos.single())
        assertEquals(1, fake.calls.size)
    }

    @Test
    fun aFailedCheckThatListsNothingIsAWarning() {
        fake.answers["derive --check"] = fake.fail("× smith.lock is not valid\n")
        assertFalse(derive())
        assertEquals("smith.lock is not valid", fake.warnings.single())
    }

    @Test
    fun theWholeListIsShownAsAPlanAndWrittenAfterAYes() {
        fake.answers["derive --dry-run"] = fake.ok("would write smith.lock\n")
        fake.ones.add(all)
        fake.plans.add(fake.yes())
        assertTrue(derive())
        assertEquals("would write smith.lock", fake.plansSeen.single())
        assertEquals(listOf("derive", "--yes"), fake.calls.last())
        assertEquals(listOf(all, "packages/web/tsconfig.json", "smith.lock"), fake.rowsSeen.first().map { it.name })
    }

    @Test
    fun aFileCanBeExplainedAndNothingChanges() {
        fake.ones.addAll(listOf("smith.lock", "explain"))
        assertFalse(derive())
        assertEquals(listOf("derive", "--explain=smith.lock"), fake.calls.last())
        assertTrue(fake.shown.last())
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aFileHeldBackIsOverwrittenOnlyAfterTheEffectIsShown() {
        fake.ones.addAll(listOf("packages/web/tsconfig.json", "replace"))
        fake.plans.add(fake.yes())
        assertTrue(derive())
        assertEquals("smith derive --replace=packages/web/tsconfig.json --yes", fake.plansSeen.single())
        assertEquals(listOf("derive", "--replace=packages/web/tsconfig.json", "--yes"), fake.calls.last())
        assertTrue(fake.infos.single().contains("Not committed"))
    }

    @Test
    fun keepingMyEditAndTheCommitBoxAddCommit() {
        fake.ones.addAll(listOf("packages/web/tsconfig.json", "adopt"))
        fake.plans.add(fake.yes(commit = true))
        assertTrue(derive())
        assertEquals(listOf("derive", "--adopt=packages/web/tsconfig.json", "--yes", "--commit"), fake.calls.last())
    }

    @Test
    fun anUntickedCommitBoxOffersTheCommitButton() {
        fake.ones.addAll(listOf("smith.lock", "eject"))
        fake.plans.add(fake.yes())
        derive()
        fake.pressLastButton()
        assertEquals("derive", committed)
    }

    @Test
    fun aStaleFileOffersTheWriteRowAndNotTheOverwriteRow() {
        fake.ones.add("smith.lock")
        derive()
        assertEquals(listOf("explain", "apply", "adopt", "eject", "reclaim"), fake.rowsSeen.last().map { it.name })
    }

    @Test
    fun cancelAtTheFileChangesNothing() {
        assertFalse(derive())
        assertTrue(fake.calls.none { "--yes" in it })
    }
}
