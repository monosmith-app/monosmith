import type { ListNode, ListResult } from '../list-view/index.ts'
import { isHeldBack } from './derive-arguments.mapper.ts'
import type { DeriveReport, DerivedFile, DeriveStatus } from './derive-report.contract.ts'

const FIXES: Record<DeriveStatus, string> = {
  stale:      'The inputs changed and nobody edited the file. Apply rewrites it.',
  missing:    'The file is gone. Apply writes it back.',
  unrecorded: 'The file is right but smith.lock does not list it. Apply records it.',
  drift:      'Someone edited the file. Explain shows the diff. Then Apply overwrites it, Adopt keeps the edit, or Eject makes the file yours.',
  diverged:   'The file differs from what the sibling projects share. Explain shows the diff. Then Apply follows the siblings, or Eject keeps the file as yours.',
  broken:     'smith cannot edit the file (a region without its markers, or a file that does not parse). Fix the file, or Eject it.',
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
      { id: `derive:${file.path}:fix`, label: `Fix: ${fixFor(file.status)}`, icon: 'lightbulb', tooltip: fixFor(file.status) },
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
      const row: ListNode = { id: 'derive:current', label: 'Derived files', description: `${report.total} up to date`, icon: 'pass', color: 'charts.green' }

      return { ...health, nodes: [...health.nodes, row] }
    }
    case 'failed': {
      const row: ListNode = { id: 'derive:failed', label: `Derived files: ${report.message ?? 'smith could not check them.'}`, icon: 'error', color: 'charts.red', tooltip: report.message }

      return { ...health, nodes: [...health.nodes, row] }
    }
    case 'pending': {
      const row: ListNode = {
        id:          'derive:pending',
        label:       'Derived files',
        description: `${report.files.length} need action`,
        icon:        'diff',
        color:       'charts.yellow',
        expanded:    true,
        children:    report.files.map((file) => toFileRow(file)),
      }
      const note = `${report.files.length} derived ${report.files.length === 1 ? 'file' : 'files'} need action`

      return { ...health, nodes: [...health.nodes, row], summary: health.summary === undefined || health.summary === 'all passed' ? note : `${health.summary}, ${note}` }
    }
  }
}
