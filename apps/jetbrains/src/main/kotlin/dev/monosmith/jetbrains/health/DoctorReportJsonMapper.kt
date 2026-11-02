package dev.monosmith.jetbrains.health

import dev.monosmith.jetbrains.jsonread.JsonFieldsMapper as Json

/** Reads `smith doctor --json` and counts its checks. */
object DoctorReportJsonMapper {
    /**
     * @param json what smith printed
     * @return the checks, in smith's order; unknown fields are ignored
     * @throws IllegalStateException when the text is not a JSON object
     */
    fun toFindings(json: String): List<DoctorFindingContract> =
        Json.records(Json.parseObject(json), "findings").map { finding ->
            DoctorFindingContract(
                check = Json.text(finding, "check"),
                ok = Json.flag(finding, "ok"),
                detail = Json.text(finding, "detail"),
                remedy = Json.text(finding, "remedy"),
                warning = Json.flag(finding, "warning"),
                fix =
                    Json.record(finding, "fix")?.let { fix ->
                        DoctorFixContract(
                            description = Json.text(fix, "description"),
                            safe = Json.flag(fix, "safe"),
                            questions =
                                Json.records(fix, "choices").map { choice ->
                                    DoctorQuestionContract(
                                        Json.text(choice, "question"),
                                        Json.records(choice, "options").map { Json.text(it, "label") },
                                    )
                                },
                        )
                    },
            )
        }

    /** Counts the checks and lays the ones that did not pass out as lines, the failed ones first. */
    fun summarize(findings: List<DoctorFindingContract>): DoctorSummaryContract {
        val problems = findings.filter { !it.ok }
        val failed = problems.filter { !it.warning }
        val warnings = problems.filter { it.warning }
        val lines =
            (failed + warnings).flatMap {
                listOf("${if (it.warning) "warning" else "failed"}: ${it.check}") +
                    (if (it.detail.isEmpty()) emptyList() else listOf("  ${it.detail}")) +
                    (if (it.remedy.isEmpty()) emptyList() else listOf("  fix: ${it.remedy}"))
            }
        return DoctorSummaryContract(findings.size - problems.size, warnings.size, failed.size, lines)
    }

    /** One sentence for a notification. */
    fun say(summary: DoctorSummaryContract): String =
        if (summary.failed == 0 && summary.warnings == 0) {
            "Doctor: all ${summary.passed} checks passed."
        } else {
            "Doctor: ${summary.failed} failed, ${summary.warnings} warnings, ${summary.passed} passed."
        }
}
