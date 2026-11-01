import { describeFailure, type SmithRun } from '../smith-run/index.ts'

/** What undoing needs from the editor. */
export interface UndoDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:     (args: string[]) => Promise<SmithRun>
  /** Shows what would be reverted (smith's plan) and answers whether to revert it. */
  confirm: (plan: string) => Promise<boolean>
}

/** How undoing ended: reverted (`lines` are what smith printed), refused with smith's words and fix, or declined by the person. */
export type UndoOutcome =
  | { kind: 'undone', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'declined' }

/**
 * Reverts the last operation smith committed (`smith undo`). It first runs
 * `--dry-run`, shows that, and only on a yes reverts with `--yes`. smith
 * itself refuses when a later commit or an uncommitted change touches the same
 * files, and says why; that is returned as the message.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function undoLastChange (deps: UndoDeps): Promise<UndoOutcome> {
  const plan = await deps.run(['undo', '--dry-run'])
  if (plan.code !== 0) return { kind: 'refused', message: describeFailure(plan) }
  if (!await deps.confirm(plan.stdout.trim())) return { kind: 'declined' }

  const done = await deps.run(['undo', '--yes'])

  return done.code === 0 ? { kind: 'undone', lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) } : { kind: 'refused', message: describeFailure(done) }
}
