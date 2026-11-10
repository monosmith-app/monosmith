import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import type { DeriveReport, DerivedFile, DeriveStatus } from './derive-report.contract.ts'

const STATUSES = new Set<string>(['stale', 'missing', 'unrecorded', 'drift', 'diverged', 'broken'])
/** `  Drift      packages/web/tsconfig.json  (tsconfig-references)  edited by hand`, as `smith derive --check` prints it. */
const FILE_LINE = /^\s+([A-Z][a-z]+)\s+(\S.*?)\s{2}\((.*?)\)\s{2}(.*)$/
const TOTAL_LINE = /^(\d+) derived files/

/**
 * Reads what `smith derive --check` printed. smith has no JSON for it yet, so
 * the lines are read as smith prints them; a line this does not know is
 * ignored, and a failure with no known line is returned as `failed` with
 * smith's own words.
 * @param run - the run of `smith derive --check`
 * @returns the state and the files that need action
 */
export function readDeriveReport (run: SmithRun): DeriveReport {
  // An older smith has no `derive` command: there is nothing to show.
  if (/derived files are off|unknown command/.test(run.stderr)) return { state: 'off', files: [], total: 0 }

  const files: DerivedFile[] = []
  let total = 0
  for (const line of run.stdout.split(/\r?\n/)) {
    const count = TOTAL_LINE.exec(line)
    if (count !== null) total = Number(count[1])
    const match = FILE_LINE.exec(line)
    const status = match?.[1]?.toLowerCase()
    if (match !== null && status !== undefined && STATUSES.has(status)) {
      files.push({ path: match[2] ?? '', status: status as DeriveStatus, renderer: match[3] ?? '', detail: match[4] ?? '' })
    }
  }
  if (files.length > 0) return { state: 'pending', files, total: total || files.length }

  // A failure with no file is also the case when only smith.lock is out of date: smith names no file, and its title and fix say what to run.
  return run.code === 0 ? { state: 'current', files: [], total } : { state: 'failed', files: [], total, message: describeFailure(run) }
}
