package dev.monosmith.jetbrains.health

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class FixReviewMapperTest {
    private val fix = DoctorFixContract("Record the vintage", false, emptyList())

    @Test
    fun theReviewIsTheFixNamedWithItsAnswersAndNeverApplied() {
        assertEquals(listOf("doctor", "--fix", "--only", "Record the vintage"), FixReviewMapper.arguments(fix, emptyList(), apply = false))
        assertEquals(
            listOf("doctor", "--fix", "--only", "Record the vintage", "--choose", "1,2", "--yes"),
            FixReviewMapper.arguments(fix, listOf(1, 2), apply = true),
        )
    }

    @Test
    fun cutsTheReviewAtTheLineThatTellsATerminalWhatToRun() {
        val stdout = "● record the vintage\n  + vintage = \"2026.11\"\n\nnot applied: run smith doctor --fix --yes\n"
        assertEquals("● record the vintage\n  + vintage = \"2026.11\"", FixReviewMapper.review(stdout))
        assertEquals("", FixReviewMapper.review("  \n"))
    }

    @Test
    fun anotherCheckStillFailingIsNotARefusal() {
        assertTrue(FixReviewMapper.stillFailing(1, listOf("fixed it", "after fixing 1 check still fails")))
        assertFalse(FixReviewMapper.stillFailing(0, listOf("after fixing")))
        assertFalse(FixReviewMapper.stillFailing(1, listOf("smith: refused")))
    }
}
