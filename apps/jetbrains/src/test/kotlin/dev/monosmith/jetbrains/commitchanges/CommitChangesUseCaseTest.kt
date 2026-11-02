package dev.monosmith.jetbrains.commitchanges

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import dev.monosmith.jetbrains.prompt.PlanAnswerContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class CommitChangesUseCaseTest {
    private val fake = FakePortsHarness()
    private val plan = "The last operation changed:\n  a.txt\nGit shows 1 more changed file:\n  b.txt\n"

    init {
        fake.answers["commit --dry-run"] = fake.ok(plan)
    }

    @Test
    fun commitsOnlyTheFilesOfTheOperationWhenTheMainButtonIsPressed() {
        fake.plans.add(fake.yes())
        fake.answers["commit --only-operation --yes"] = fake.ok("committed abc123\n")
        assertTrue(CommitChangesUseCase.run(fake.ports))
        assertEquals(plan.trim(), fake.plansSeen.single())
        assertEquals(listOf("commit", "--only-operation", "--yes"), fake.calls.last())
        assertEquals("committed abc123", fake.infos.single())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun commitsEverythingGitShowsOnlyWhenTheOtherButtonIsPressed() {
        fake.plans.add(PlanAnswerContract(false, emptyList(), 0))
        assertTrue(CommitChangesUseCase.run(fake.ports))
        assertEquals(listOf("commit", "--all", "--yes"), fake.calls.last())
    }

    @Test
    fun theWordsOfTheOperationNameTheCommit() {
        fake.answers["commit add app web --dry-run"] = fake.ok(plan)
        fake.plans.add(fake.yes())
        CommitChangesUseCase.run(fake.ports, "add app web")
        assertEquals(listOf("commit", "add", "app", "web", "--dry-run"), fake.calls.first())
        assertEquals(listOf("commit", "add", "app", "web", "--only-operation", "--yes"), fake.calls.last())
    }

    @Test
    fun cancelCommitsNothing() {
        assertFalse(CommitChangesUseCase.run(fake.ports))
        assertTrue(fake.calls.none { "--yes" in it })
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aRefusedPlanIsAWarningWithSmithsWords() {
        fake.answers["commit --dry-run"] = fake.fail("✖ there is nothing to commit\n")
        assertFalse(CommitChangesUseCase.run(fake.ports))
        assertEquals("Nothing was committed. there is nothing to commit", fake.warnings.single())
    }

    @Test
    fun aFailedCommitIsAWarningAndStillTellsTheViewsTheTreeMayHaveChanged() {
        fake.plans.add(fake.yes())
        fake.answers["commit --only-operation --yes"] = fake.fail("✖ git refused the commit\n")
        assertFalse(CommitChangesUseCase.run(fake.ports))
        assertTrue(fake.warnings.single().startsWith("Nothing was committed."))
        assertEquals(1, fake.changedCount)
    }
}
