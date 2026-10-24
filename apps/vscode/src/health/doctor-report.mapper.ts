import type { DoctorReport } from './doctor-report.contract.ts'

/** What a doctor report comes to. */
export interface DoctorSummary {
  passed:   number
  warnings: number
  failed:   number
  /** The lines to show, the checks that did not pass first. */
  lines:    string[]
}

/** Counts the checks and lays the ones that did not pass out as lines for the output. */
export function summarizeDoctor (report: DoctorReport): DoctorSummary {
  const findings = report.findings ?? []
  const problems = findings.filter((finding) => !finding.ok)
  const failed = problems.filter((finding) => finding.warning !== true)
  const warnings = problems.filter((finding) => finding.warning === true)
  const lines = [...failed, ...warnings].flatMap((finding) => [
    `${finding.warning === true ? 'warning' : 'failed'}: ${finding.check}`,
    ...(finding.detail === undefined ? [] : [`  ${finding.detail}`]),
    ...(finding.remedy === undefined ? [] : [`  fix: ${finding.remedy}`]),
  ])

  return { passed: findings.length - problems.length, warnings: warnings.length, failed: failed.length, lines }
}

/** One sentence for a notification. */
export function sayDoctor (summary: DoctorSummary): string {
  if (summary.failed === 0 && summary.warnings === 0) return `Doctor: all ${summary.passed} checks passed.`
  const parts = [`${summary.failed} failed`, `${summary.warnings} warnings`, `${summary.passed} passed`]

  return `Doctor: ${parts.join(', ')}.`
}
