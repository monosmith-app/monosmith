package dev.monosmith.jetbrains.health

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

class DoctorReportJsonMapperTest {
    private val json =
        """
        {"findings":[
          {"check":"git is installed","ok":true},
          {"check":"the cache is ignored","ok":false,"detail":"not in .gitignore","remedy":"add it",
           "fix":{"description":"ignore the cache","safe":true}},
          {"check":"defaults","ok":false,"warning":true,"detail":"d","remedy":"r",
           "fix":{"description":"Record the vintage","safe":false,
                  "choices":[{"question":"What do you want?","options":[{"label":"keep"},{"label":"move"}]}]}},
          {"check":"lockfile","ok":false}
        ]}
        """.trimIndent()

    @Test
    fun readsTheChecksAndTheirFixes() {
        val findings = DoctorReportJsonMapper.toFindings(json)
        assertEquals(4, findings.size)
        assertNull(findings.first().fix)
        assertTrue(findings[1].fix!!.safe)
        val asked = findings[2].fix!!
        assertEquals("What do you want?", asked.questions.single().question)
        assertEquals(listOf("keep", "move"), asked.questions.single().options)
    }

    @Test
    fun countsAndLaysOutTheChecksThatDidNotPassFailedFirst() {
        val summary = DoctorReportJsonMapper.summarize(DoctorReportJsonMapper.toFindings(json))
        assertEquals(1, summary.passed)
        assertEquals(2, summary.failed)
        assertEquals(1, summary.warnings)
        assertEquals(
            listOf(
                "failed: the cache is ignored",
                "  not in .gitignore",
                "  fix: add it",
                "failed: lockfile",
                "warning: defaults",
                "  d",
                "  fix: r",
            ),
            summary.lines,
        )
    }

    @Test
    fun saysOneSentence() {
        assertEquals("Doctor: all 3 checks passed.", DoctorReportJsonMapper.say(DoctorSummaryContract(3, 0, 0, emptyList())))
        assertEquals("Doctor: 1 failed, 2 warnings, 5 passed.", DoctorReportJsonMapper.say(DoctorSummaryContract(5, 2, 1, emptyList())))
    }
}
