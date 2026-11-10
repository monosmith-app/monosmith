import type { ListNode, ListResult } from '../list-view/index.ts'
import { isHeldBack } from './derive-arguments.mapper.ts'
import type { DeriveReport, DerivedFile, DeriveStatus } from './derive-report.contract.ts'
import { say } from '../words/index.ts'

const FIXES: Record<DeriveStatus, string> = {
  stale:      say('vscode.derive-files.the-inputs-changed-and-nobody-edited'),
  missing:    say('vscode.derive-files.the-file-is-gone-apply-writes'),
  unrecorded: say('vscode.derive-files.the-file-is-right-but-smith'),
  drift:      say('vscode.derive-files.someone-edited-the-file-explain-shows'),
  diverged:   say('vscode.derive-files.the-file-differs-from-what-the'),
  broken:     say('vscode.derive-files.smith-cannot-edit-the-file-a'),
}

const ICONS: Record<DeriveStatus, string> = {
  stale: 'sync', missing: 'file-add', unrecorded: 'bookmark', drift: 'diff', diverged: 'git-compare', broken: 'error',
}

/**
 * The fix for a state, in a sentence for a person.
 * @param status - the state of the file
 */
export function fixFor (status: DeriveStatus): string {
  return FIXES[status]
}

/**
 * What a menu sees as `viewItem`: a file smith can write (`derived.file`) or
 * one it holds back until the person chooses (`derived.held`).
 * @param status - the state of the file
 */
export function contextFor (status: DeriveStatus): 'derived.file' | 'derived.held' {
  return isHeldBack(status) ? 'derived.held' : 'derived.file'
}

function toFileRow (file: DerivedFile): ListNode {
  const tooltip = `${file.path}\n${file.status}: ${file.detail}\n${fixFor(file.status)}`

  return {
    id:          `derive:${file.path}`,
    label:       file.path,
    description: `${file.status} · ${file.renderer}`,
    icon:        ICONS[file.status],
    color:       file.status === 'broken' ? 'charts.red' : 'charts.yellow',
    context:     contextFor(file.status),
    data:        file.path,
    tooltip,
    children:    [
      ...(file.detail === '' ? [] : [{ id: `derive:${file.path}:detail`, label: file.detail, icon: 'info', tooltip: file.detail }]),
      { id: `derive:${file.path}:fix`, label: say('vscode.derive-files.fix', { fixFor: fixFor(file.status) }), icon: 'lightbulb', tooltip: fixFor(file.status) },
    ],
  }
}

/**
 * Adds the derived files to the rows of the Health view: a "Derived files"
 * row, with a row for each file that drifted or diverges, its state, and the
 * fix. Nothing is added when the feature is off.
 * @param health - the rows the doctor produced
 * @param report - what `smith derive --check` found
 * @returns the rows and a summary that counts the derived files that need action
 */
export function addDeriveRows (health: ListResult, report: DeriveReport): ListResult {
  switch (report.state) {
    case 'off': {
      return health
    }
    case 'current': {
      const row: ListNode = { id: 'derive:current', label: say('vscode.derive-files.derived-files'), description: `${report.total} up to date`, icon: 'pass', color: 'charts.green' }

      return { ...health, nodes: [...health.nodes, row] }
    }
    case 'failed': {
      const row: ListNode = { id: 'derive:failed', label: say('vscode.derive-files.derived-files-2', { part: report.message ?? say('vscode.derive-files.smith-could-not-check-them') }), icon: 'error', color: 'charts.red', tooltip: report.message }

      return { ...health, nodes: [...health.nodes, row] }
    }
    case 'pending': {
      const row: ListNode = {
        id:          'derive:pending',
        label:       say('vscode.derive-files.derived-files'),
        description: say('vscode.derive-files.need-action', { length: report.files.length }),
        icon:        'diff',
        color:       'charts.yellow',
        expanded:    true,
        children:    report.files.map((file) => toFileRow(file)),
      }
      const note = say('vscode.derive-files.derived-need-action', { length: report.files.length, part: say(report.files.length === 1 ? 'vscode.agent-install.file' : 'vscode.agent-install.files') })

      return { ...health, nodes: [...health.nodes, row], summary: health.summary === undefined || health.summary === say('doctor.group.passed') ? note : `${health.summary}, ${note}` }
    }
  }
}
