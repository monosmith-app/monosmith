package dev.monosmith.jetbrains.releaseflow

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ReleaseWorkspaceUseCaseTest {
    private val fake = FakePortsHarness()
    private var preview = 0
    private var restarted = 0

    private fun release() = ReleaseWorkspaceUseCase.run(fake.ports, { preview++ }, { restarted++ })

    private fun planned(unchanged: String = "[]") =
        fake.ok(
            """{"results":[{"project":"a","next":"1.1.0","level":"minor","current":"1.0.0"},""" +
                """{"project":"b","next":"1.0.1","level":"patch","current":"1.0.0"}],"unchanged":$unchanged}""",
        )

    @Test
    fun publishesOnlyAfterTheLastQuestionAndWithTheHoldAndTheBumpTheyChose() {
        fake.answers["release --dry-run --json"] = planned()
        fake.answers["release --dry-run"] = fake.ok("a 1.0.0 -> 1.1.0\nb 1.0.0 -> 1.0.1\n")
        fake.answers["release --yes"] = fake.ok("released a\n")
        fake.plans.addAll(listOf(fake.yes(), fake.yes()))
        fake.manies.add(listOf("b"))
        fake.ones.addAll(listOf("a", "major", ""))
        assertTrue(release())
        assertEquals(listOf("release", "--yes", "--hold", "b", "--bump", "a=major"), fake.calls.last())
        assertTrue(fake.plansSeen.last().contains("Held back: b."))
        assertTrue(fake.plansSeen.last().contains("Changed bumps: a=major."))
        assertEquals("released a", fake.infos.single())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun plainChoicesReleaseWhatTheCommitsDecided() {
        fake.answers["release --dry-run --json"] = planned()
        fake.answers["release --dry-run"] = fake.ok("a, b\n")
        fake.plans.addAll(listOf(fake.yes(), fake.yes()))
        fake.manies.add(emptyList())
        fake.ones.add("")
        assertTrue(release())
        assertEquals(listOf("release", "--yes"), fake.calls.last())
    }

    @Test
    fun cancelAtTheLastQuestionPublishesNothing() {
        fake.answers["release --dry-run --json"] = planned()
        fake.plans.add(fake.yes())
        fake.manies.add(emptyList())
        fake.ones.add("")
        assertFalse(release())
        assertTrue(fake.calls.none { "--yes" in it })
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aPackageWithoutAReleasingCommitIsReleasedOnlyWhenThePersonTicksItAndNamesALevel() {
        fake.answers["release --dry-run --json"] = fake.ok("""{"results":[],"unchanged":["c"]}""")
        fake.answers["release --dry-run"] = fake.ok("Nothing to release.\n")
        fake.manies.add(listOf("c"))
        fake.ones.add("patch")
        fake.plans.add(fake.yes())
        assertTrue(release())
        assertEquals(listOf("release", "--yes", "--bump", "c=patch", "--force"), fake.calls.last())
        assertTrue(fake.plansSeen.single().contains("c: no commit releases it."))
    }

    @Test
    fun whenNothingReleasesTheNoticeOffersThePreviewAndTheReleaseWithoutACommit() {
        fake.answers["release --dry-run --json"] = fake.ok("""{"results":[],"unchanged":["c"]}""")
        fake.answers["release --dry-run"] = fake.ok("Nothing to release.\n")
        fake.manies.add(emptyList())
        assertFalse(release())
        assertTrue(fake.infos.single().startsWith("Nothing to release."))
        val buttons = fake.buttons.single()
        assertEquals(listOf("Open Preview", "Without a commit…"), buttons.map { it.label })
        buttons.forEach { it.action() }
        assertEquals(1, preview)
        assertEquals(1, restarted)
        assertTrue(fake.calls.none { "--yes" in it })
    }

    @Test
    fun aRefusedPlanIsAWarningAndAsksNothing() {
        fake.answers["release --dry-run"] = fake.fail("✖ the branch is not main\n→ switch to main\n")
        assertFalse(release())
        assertEquals("Nothing was released. the branch is not main Try: switch to main", fake.warnings.single())
        assertEquals(emptyList(), fake.plansSeen)
    }

    @Test
    fun aPlanItCannotReadIsATextThatSaysToUpdateThePlugin() {
        fake.answers["release --dry-run --json"] = fake.ok("not json")
        assertFalse(release())
        assertTrue(fake.warnings.single().contains("Update the plugin"))
    }
}
