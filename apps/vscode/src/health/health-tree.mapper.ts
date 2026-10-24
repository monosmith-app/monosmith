import type { ListNode, ListResult } from '../list-view/index.ts'
import type { DoctorFinding, DoctorReport } from './doctor-report.contract.ts'
import type { FixPreview } from './fix-preview.mapper.ts'

/**
 * Lays a doctor report out as rows: the failed checks first, then the
 * warnings, each with the problem and the remedy under it, then the fixes
 * smith proposes, and the checks that passed, closed.
 * @param report - what `doctor` answered
 * @param preview - the fixes `doctor.fixes` proposes, if any
 */
export function toHealthRows (report: DoctorReport, preview?: FixPreview): ListResult {
  const findings = report.findings ?? []
  const problems = findings.filter((finding) => !finding.ok)
  const failed = problems.filter((finding) => finding.warning !== true)
  const warnings = problems.filter((finding) => finding.warning === true)
  const passed = findings.filter((finding) => finding.ok)
  if (findings.length === 0) return { nodes: [], message: 'smith reported no checks.' }

  const nodes: ListNode[] = [
    ...(problems.length === 0 ? [{ id: 'health:ok', label: `All ${passed.length} checks passed`, icon: 'pass', color: 'charts.green' }] : []),
    ...failed.map((finding, index) => toProblemRow(finding, `failed:${index}`, 'error', 'charts.red')),
    ...warnings.map((finding, index) => toProblemRow(finding, `warning:${index}`, 'warning', 'charts.yellow')),
    ...(preview === undefined
      ? []
      : [{
          id:          'health:fixes',
          label:       'Fixes available',
          description: String(preview.titles.length),
          icon:        'wrench',
          expanded:    true,
          children:    preview.titles.map((title, index) => ({ id: `health:fix:${index}`, label: title, icon: 'lightbulb', tooltip: preview.text })),
        }]),
    ...(passed.length > 0 && problems.length > 0
      ? [{
          id:          'health:passed',
          label:       'Passed',
          description: String(passed.length),
          icon:        'pass',
          children:    passed.map((finding, index) => ({ id: `health:passed:${index}`, label: finding.check, icon: 'pass', color: 'charts.green' })),
        }]
      : []),
  ]

  return { nodes, summary: summarize(failed.length, warnings.length) }
}

function toProblemRow (finding: DoctorFinding, id: string, icon: string, color: string): ListNode {
  const detail = finding.detail?.trim()
  const remedy = finding.remedy?.trim()
  const children: ListNode[] = [
    ...(detail === undefined || detail === '' ? [] : [{ id: `${id}:problem`, label: detail, icon: 'info', tooltip: detail }]),
    ...(remedy === undefined || remedy === '' ? [] : [{ id: `${id}:remedy`, label: remedy, icon: 'lightbulb', tooltip: `Fix: ${remedy}` }]),
  ]

  return {
    id:       `health:${id}`,
    label:    finding.check,
    icon,
    color,
    expanded: true,
    tooltip:  [finding.check, detail, remedy === undefined ? undefined : `Fix: ${remedy}`].filter(Boolean).join('\n'),
    children,
  }
}

function summarize (failed: number, warnings: number): string {
  if (failed === 0 && warnings === 0) return 'all passed'

  return [failed > 0 ? `${failed} failed` : '', warnings > 0 ? `${warnings} ${warnings === 1 ? 'warning' : 'warnings'}` : ''].filter(Boolean).join(', ')
}
