import type { ListNode, ListResult } from '../../shared/list-view/index.ts'
import type { DoctorFinding, DoctorReport } from './doctor-report.contract.ts'
import type { FixPreview } from './fix-preview.mapper.ts'
import { say, type WordKey } from '../../shared/words/index.ts'

/**
 * Lays a doctor report out as the Health view of the terminal does: the checks grouped by
 * state (Failed, Warnings, then Passed, closed), each heading with its count. A check says its
 * project and its file, and `fix` when smith can fix it; under it come the problem, the remedy
 * and the fix, which is reviewed, applied and checked again in the editor. The fixes smith
 * proposes for the whole workspace follow the warnings.
 * @param report - what `doctor` answered
 * @param preview - the fixes `doctor.fixes` proposes, if any
 * @returns the rows, with the counts beside the title
 */
export function toHealthRows (report: DoctorReport, preview?: FixPreview): ListResult {
  const findings = report.findings ?? []
  const problems = findings.filter((finding) => !finding.ok)
  const failed = problems.filter((finding) => finding.warning !== true)
  const warnings = problems.filter((finding) => finding.warning === true)
  const passed = findings.filter((finding) => finding.ok)
  if (findings.length === 0) return { nodes: [], message: say('vscode.health.smith-reported-no-checks') }

  const nodes: ListNode[] = [
    ...(problems.length === 0 ? [{ id: 'health:ok', label: say('vscode.health.all-checks-passed', { length: passed.length }), icon: 'pass', color: 'charts.green' }] : []),
    ...stateGroup('failed', 'doctor.state.failed', failed, 'error', 'charts.red'),
    ...stateGroup('warning', 'doctor.state.warnings', warnings, 'warning', 'charts.yellow'),
    ...(preview === undefined
      ? []
      : [{
          id:          'health:fixes',
          label:       say('vscode.health.fixes-available'),
          description: String(preview.titles.length),
          icon:        'wrench',
          expanded:    true,
          children:    preview.titles.map((title, index) => ({ id: `health:fix:${index}`, label: title, icon: 'lightbulb', tooltip: preview.text })),
        }]),
    ...(passed.length > 0 && problems.length > 0
      ? [{
          id:          'health:passed',
          label:       say('doctor.state.passed'),
          description: String(passed.length),
          icon:        'pass',
          color:       'charts.green',
          children:    passed.map((finding, index) => ({ id: `health:passed:${index}`, label: finding.check, icon: 'pass', color: 'charts.green', description: whereOf(finding) })),
        }]
      : []),
  ]

  return { nodes, summary: summarize(failed.length, warnings.length, passed.length) }
}

/** A heading of one state with its checks under it, open; nothing when no check is in that state. */
function stateGroup (state: string, label: WordKey, findings: readonly DoctorFinding[], icon: string, color: string): ListNode[] {
  if (findings.length === 0) return []

  return [{
    id:          `health:${state}`,
    label:       say(label),
    description: String(findings.length),
    icon,
    color,
    expanded:    true,
    children:    findings.map((finding, index) => toProblemRow(finding, `${state}:${index}`, icon, color)),
  }]
}

/**
 * The file a check is about, as the terminal's Health list names it: the first file its fix
 * writes, else the first word of its problem or detail that looks like a path.
 * @param finding - the check
 * @returns the path, or an empty text when none is named
 */
export function fileOf (finding: DoctorFinding): string {
  const written = finding.fix?.files?.find((file) => file.path !== '')?.path
  if (written !== undefined) return written
  const fields = [finding.problem ?? '', finding.detail ?? ''].flatMap((text) => text.split(/\s+/))
  // A sentence's full stop goes; the dot that starts ".gitignore" stays.
  const paths = fields.map((field) => field.replace(/^[`'"()[\],;:]+/, '').replace(/[`'"()[\],;:.]+$/, ''))

  return paths.find((path) => looksLikePath(path)) ?? ''
}

/** True for a word that names a file: its last part has a name and an extension of letters, and it is no URL, flag or version. */
function looksLikePath (path: string): boolean {
  if (path === '' || path.startsWith('-') || path.includes('://') || /[<>{}=*]/.test(path)) return false
  const last = path.slice(path.lastIndexOf('/') + 1)

  return /^(?:[\w.-]*\w)?\.[a-z]{2,}$/i.test(last) && !/^\d+(?:\.\d+)+$/.test(last) && !/^[a-z]\.[a-z]$/i.test(last)
}

/** `web · package-lock.json`: the project (or the workspace) and the file of a check. */
function whereOf (finding: DoctorFinding): string {
  const project = finding.project === undefined || finding.project === '' ? say('vscode.health.workspace') : finding.project
  const file = fileOf(finding)

  return file === '' ? project : say('doctor.row.where', { project, file })
}

/**
 * The command line that fixes one check in a terminal, as the fix in the editor runs it.
 * @param finding - the check with a fix
 * @returns `smith doctor --fix --only '<fix>'`
 */
export function fixCommandLine (finding: DoctorFinding): string {
  return `smith doctor --fix --only '${finding.fix?.description ?? finding.check}'`
}

function toProblemRow (finding: DoctorFinding, id: string, icon: string, color: string): ListNode {
  const detail = (finding.problem ?? finding.detail)?.trim()
  const remedy = finding.remedy?.trim()
  const where = whereOf(finding)
  const children: ListNode[] = [
    ...(detail === undefined || detail === '' ? [] : [{ id: `${id}:problem`, label: firstClause(detail), icon: 'info', tooltip: detail }]),
    ...(remedy === undefined || remedy === '' ? [] : [{ id: `${id}:remedy`, label: firstClause(remedy), icon: 'lightbulb', tooltip: say('vscode.health.fix-2', { remedy }) }]),
    // A fix runs in the editor: smith shows what it would change and asks first, so no remedy has to be typed into a terminal.
    ...(finding.fix === undefined
      ? []
      : [{
          id:          `${id}:fix`,
          label:       say('vscode.health.fix-3', { description: finding.fix.description }),
          description: steps(),
          icon:        'wrench',
          context:     'health.fix',
          command:     'monosmith.health.fixFinding',
          data:        finding.check,
          tooltip:     `${say('vscode.health.shows-what-it-changes-and-asks')}\n${say('tui.detail.or_in_a_terminal_smith_doctor', { check: finding.fix.description })}`,
        }, {
          id:          `${id}:same`,
          label:       say('releasepage.detail.same_as'),
          description: `$ ${fixCommandLine(finding)}`,
          icon:        'terminal',
          tooltip:     fixCommandLine(finding),
        }]),
  ]

  return {
    id:          `health:${id}`,
    label:       finding.check,
    description: finding.fix === undefined ? where : `${where} · ${say('doctor.row.fix')}`,
    icon,
    color,
    expanded:    true,
    tooltip:     [finding.check, where, detail, remedy === undefined ? undefined : say('vscode.health.fix-2', { remedy })].filter(Boolean).join('\n'),
    children,
  }
}

/** `1 Review ── 2 Apply ── 3 Check again`: the steps of a fix, as the terminal draws them. */
function steps (): string {
  return [say('healthfix.step.review'), say('healthfix.step.apply'), say('healthfix.step.check')].map((step, index) => `${index + 1} ${step}`).join(' ── ')
}

/** The part of a sentence up to its first `;` or `. `, so a narrow row shows the part that matters; the tooltip has all of it. */
function firstClause (text: string): string {
  return /^(.+?)(?:;|\. )/.exec(text)?.[1] ?? text
}

function summarize (failed: number, warnings: number, passed: number): string {
  if (failed === 0 && warnings === 0) return say('doctor.group.passed')

  return [failed > 0 ? `${failed} failed` : '', warnings > 0 ? `${warnings} ${say(warnings === 1 ? 'vscode.health.warning' : 'vscode.health.warnings')}` : '', `${passed} ${say('vscode.health.passed-word')}`].filter(Boolean).join(' · ')
}
