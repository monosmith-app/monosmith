package dev.monosmith.jetbrains.planapply

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class PlanThenApplyUseCaseTest {
    private val fake = FakePortsHarness()

    private fun step(offerCommit: Boolean = true) =
        PlanStepContract(
            title = "remove web?",
            planArgs = listOf("remove", "web", "--dry-run"),
            header = "Nothing was changed yet.",
            okText = "Remove",
            offerCommit = offerCommit,
            apply = { listOf("remove", "web", "--keep-copy", "--yes") },
            done = "web is removed",
            operation = "remove web",
        )

    @Test
    fun showsThePlanFirstAndNeverCommitsUnlessTheBoxIsTicked() {
        fake.answers["remove web --dry-run"] = fake.ok("would move apps/web\n")
        fake.plans.add(fake.yes())
        assertTrue(PlanThenApplyUseCase.run(fake.ports, step()) {})
        assertEquals(listOf("would move apps/web"), fake.plansSeen)
        assertEquals(listOf("remove", "web", "--keep-copy", "--yes"), fake.calls.last())
        assertTrue(fake.infos.single().contains("Not committed"))
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun aTickedBoxAddsCommitAndOffersNoCommitButton() {
        fake.plans.add(fake.yes(commit = true))
        PlanThenApplyUseCase.run(fake.ports, step()) {}
        assertEquals("--commit", fake.calls.last().last())
        assertTrue(fake.infos.single().contains("Committed as one change"))
        assertEquals(emptyList(), fake.buttons.single())
    }

    @Test
    fun anUncommittedChangeOffersAButtonThatStartsTheCommit() {
        fake.plans.add(fake.yes())
        var operation = ""
        PlanThenApplyUseCase.run(fake.ports, step()) { operation = it }
        fake.buttons
            .single()
            .single()
            .action()
        assertEquals("remove web", operation)
    }

    @Test
    fun cancelLeavesEverythingAsItWas() {
        assertFalse(PlanThenApplyUseCase.run(fake.ports, step()) {})
        assertEquals(1, fake.calls.size)
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aRefusedPlanIsATextAndNoQuestion() {
        fake.answers["remove web --dry-run"] = fake.fail("× the tree is not clean\n→ commit first\n")
        assertFalse(PlanThenApplyUseCase.run(fake.ports, step()) {})
        assertEquals("Nothing was changed. the tree is not clean Try: commit first", fake.warnings.single())
        assertEquals(emptyList(), fake.plansSeen)
    }

    @Test
    fun aFailedChangeWarnsAndKeepsTheCommitBoxOut() {
        fake.answers["remove web --keep-copy"] = fake.fail("× locked")
        fake.plans.add(fake.yes(commit = true))
        assertFalse(PlanThenApplyUseCase.run(fake.ports, step()) {})
        assertTrue(fake.warnings.single().startsWith("Nothing was changed. locked"))
    }

    @Test
    fun aCommandWithoutACommitFlagNeverGetsOne() {
        fake.plans.add(fake.yes(commit = true))
        PlanThenApplyUseCase.run(fake.ports, step(offerCommit = false)) {}
        assertFalse("--commit" in fake.calls.last())
        assertTrue(fake.infos.single().endsWith("Nothing is committed."))
    }
}
