package dev.monosmith.jetbrains.health

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import dev.monosmith.jetbrains.smithrun.SmithRunContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class HealthFixUseCaseTest {
    private val fake = FakePortsHarness()
    private val fix =
        DoctorFixContract("Record the vintage", false, listOf(DoctorQuestionContract("Which vintage?", listOf("2026.11", "2026.12"))))
    private val finding = DoctorFindingContract("vintage", false, "none recorded", "record one", false, fix)
    private val review = "● record the vintage\n  + vintage = \"2026.12\"\n\nnot applied: run smith doctor --fix --yes\n"

    @Test
    fun asksTheQuestionsShowsSmithsReviewAndAppliesOnlyAfterAYes() {
        fake.answers["doctor --fix --only Record the vintage --choose 2"] = fake.ok(review)
        fake.answers["doctor --fix --only Record the vintage --choose 2 --yes"] = fake.ok("fixed\n")
        fake.ones.add("1")
        fake.plans.add(fake.yes())
        assertTrue(HealthUseCase.runFix(fake.ports, finding))
        val rows = fake.rowsSeen.single()
        assertEquals("(suggested)", rows.first().description)
        assertEquals("● record the vintage\n  + vintage = \"2026.12\"", fake.plansSeen.single())
        assertEquals(listOf("doctor", "--fix", "--only", "Record the vintage", "--choose", "2", "--yes"), fake.calls.last())
        assertEquals("The fix is applied. Review it with the Commit window.", fake.infos.single())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun cancelAtAQuestionRunsNothing() {
        assertFalse(HealthUseCase.runFix(fake.ports, finding))
        assertEquals(emptyList(), fake.calls)
    }

    @Test
    fun cancelAtTheReviewChangesNothing() {
        fake.answers["doctor --fix --only"] = fake.ok(review)
        fake.ones.add("0")
        assertFalse(HealthUseCase.runFix(fake.ports, finding))
        assertTrue(fake.calls.none { "--yes" in it })
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aCheckWithoutAFixSaysSoAndRunsNothing() {
        assertFalse(HealthUseCase.runFix(fake.ports, finding.copy(fix = null)))
        assertEquals("smith has no fix for this check. record one", fake.infos.single())
        assertEquals(emptyList(), fake.calls)
    }

    @Test
    fun anEmptyReviewIsAWarningWithSmithsWords() {
        fake.answers["doctor --fix --only"] = fake.fail("✖ this fix needs a terminal\n")
        fake.ones.add("0")
        assertFalse(HealthUseCase.runFix(fake.ports, finding))
        assertEquals("Nothing was changed. this fix needs a terminal", fake.warnings.single())
    }

    @Test
    fun anotherCheckStillFailingAfterTheFixIsNotARefusal() {
        fake.answers["doctor --fix --only Record the vintage --choose 1"] = fake.ok(review)
        fake.answers["doctor --fix --only Record the vintage --choose 1 --yes"] =
            SmithRunContract(1, "fixed\nafter fixing 1 check still fails\n", "")
        fake.ones.add("0")
        fake.plans.add(fake.yes())
        assertTrue(HealthUseCase.runFix(fake.ports, finding))
        assertEquals("The fix is applied. Another check still fails.", fake.infos.single())
    }

    @Test
    fun theSafeFixesAreShownAsDiffsAndAppliedAfterAYes() {
        fake.answers["doctor --fix"] = fake.ok("● pin the JDK\n  + java = 21\n\nnot applied: run smith doctor --fix --yes\n")
        fake.plans.add(fake.yes())
        assertTrue(HealthUseCase.runFixSafe(fake.ports))
        assertEquals("● pin the JDK\n  + java = 21", fake.plansSeen.single())
        assertEquals(listOf("doctor", "--fix", "--yes"), fake.calls.last())
        assertTrue(fake.shown.last())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun withNoSafeFixNothingIsAskedAndNothingChanges() {
        fake.answers["doctor --fix"] = fake.ok("")
        assertFalse(HealthUseCase.runFixSafe(fake.ports))
        assertTrue(fake.infos.single().startsWith("smith has no safe fix to apply."))
        assertEquals(emptyList(), fake.plansSeen)
    }

    @Test
    fun cancelAtTheSafeFixesChangesNothing() {
        fake.answers["doctor --fix"] = fake.ok("● pin the JDK\n")
        assertFalse(HealthUseCase.runFixSafe(fake.ports))
        assertTrue(fake.calls.none { "--yes" in it })
    }
}
