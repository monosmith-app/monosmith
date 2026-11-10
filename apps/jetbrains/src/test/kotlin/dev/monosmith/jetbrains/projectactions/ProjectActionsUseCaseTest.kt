package dev.monosmith.jetbrains.projectactions

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import dev.monosmith.jetbrains.prompt.PlanAnswerContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ProjectActionsUseCaseTest {
    private val fake = FakePortsHarness()
    private var committed = ""
    private val projects =
        """
        [{"name":"web","dir":"apps/web","kind":"app","language":"node","tasks":["build"],"deps":["utils"]},
         {"name":"utils","dir":"libs/utils","kind":"lib","language":"node","tasks":["build"],"deps":[]}]
        """.trimIndent()

    init {
        fake.answers["projects --json"] = fake.ok(projects)
    }

    @Test
    fun removeShowsThePlanAndKeepsACopyByDefault() {
        fake.answers["remove web --dry-run"] = fake.ok("would move apps/web to .smith/removed/web\n")
        fake.plans.add(fake.yes())
        assertTrue(ProjectActionsUseCase.runRemove(fake.ports, { committed = it }, "web"))
        assertEquals("would move apps/web to .smith/removed/web", fake.plansSeen.single())
        assertEquals(listOf("remove", "web", "--keep-copy", "--yes"), fake.calls.last())
        assertEquals("", committed)
    }

    @Test
    fun theOtherButtonDeletesTheFolderAndAFolderIsNeverDeletedBecauseOfTheMainButton() {
        fake.plans.add(PlanAnswerContract(false, emptyList(), 0))
        assertTrue(ProjectActionsUseCase.runRemove(fake.ports, {}, "web"))
        assertEquals(listOf("remove", "web", "--delete", "--yes"), fake.calls.last())
    }

    @Test
    fun theCommitBoxAddsCommitAndAnUntickedBoxOffersTheButton() {
        fake.plans.addAll(listOf(fake.yes(commit = true), fake.yes()))
        ProjectActionsUseCase.runRemove(fake.ports, {}, "web")
        assertEquals("--commit", fake.calls.last().last())
        ProjectActionsUseCase.runRemove(fake.ports, { committed = it }, "web")
        fake.pressLastButton()
        assertEquals("remove web", committed)
    }

    @Test
    fun cancelRemovesNothing() {
        assertFalse(ProjectActionsUseCase.runRemove(fake.ports, {}, "web"))
        assertTrue(fake.calls.none { "--yes" in it })
    }

    @Test
    fun withNoProjectGivenItAsksWhichAndOffersEachProject() {
        fake.ones.add("utils")
        ProjectActionsUseCase.runRemove(fake.ports, {}, null)
        assertEquals(listOf("web", "utils"), fake.rowsSeen.single().map { it.name })
        assertEquals(listOf("remove", "utils", "--dry-run"), fake.calls.last())
    }

    @Test
    fun aWorkspaceItCannotReadIsAWarningAndNothingElseRuns() {
        fake.answers["projects --json"] = fake.fail("× smith.toml:3:9 is not valid\n")
        assertFalse(ProjectActionsUseCase.runRemove(fake.ports, {}, "web"))
        assertTrue(fake.warnings.single().startsWith("Nothing was changed."))
        assertEquals(1, fake.calls.size)
    }

    @Test
    fun theKindIsAskedWithWhatEachMeansThenPlannedThenChanged() {
        fake.ones.add("lib")
        fake.answers["kind web lib --dry-run"] = fake.ok("would move apps/web to libs/web\n")
        fake.plans.add(fake.yes())
        assertTrue(ProjectActionsUseCase.runChangeKind(fake.ports, {}, "web"))
        assertEquals(listOf("lib", "package"), fake.rowsSeen.single().map { it.name })
        assertEquals("would move apps/web to libs/web", fake.plansSeen.single())
        assertEquals(listOf("kind", "web", "lib", "--yes"), fake.calls.last())
        assertTrue(fake.infos.single().startsWith("web is now a lib."))
    }

    @Test
    fun aFailedKindChangeIsAWarning() {
        fake.ones.add("package")
        fake.answers["kind web package --dry-run"] = fake.fail("× web has no version\n→ give it one\n")
        assertFalse(ProjectActionsUseCase.runChangeKind(fake.ports, {}, "web"))
        assertEquals("Nothing was changed. web has no version Try: give it one", fake.warnings.single())
    }

    @Test
    fun unlinkOffersOnlyTheLibrariesTheProjectUsesAndNeverCommits() {
        fake.ones.add("utils")
        fake.plans.add(fake.yes())
        assertTrue(ProjectActionsUseCase.runUnlink(fake.ports, "web") {})
        assertEquals(listOf("utils"), fake.rowsSeen.single().map { it.name })
        assertEquals("smith uninstall -p web workspace:utils", fake.plansSeen.single())
        assertEquals(listOf("uninstall", "-p", "web", "workspace:utils", "--yes"), fake.calls.last())
        assertTrue(fake.infos.single().startsWith("utils is unlinked from web."))
    }

    @Test
    fun aProjectThatUsesNoLibraryHasNothingToUnlink() {
        assertFalse(ProjectActionsUseCase.runUnlink(fake.ports, "utils") {})
        assertEquals("utils depends on no library of this workspace.", fake.infos.single())
        assertEquals(1, fake.calls.size)
    }
}
