import { describeFailure, type SmithRun } from '../smith-run/index.ts'

/** What committing a change after the fact needs from the editor. */
export interface CommitChangesDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:     (args: string[]) => Promise<SmithRun>
  /** Shows the files that would be committed (smith's plan) and answers whether to commit them. */
  confirm: (plan: string) => Promise<boolean>
}

/** How committing ended: committed (`lines` are what smith printed), refused with smith's words, or declined by the person. */
export type CommitChangesOutcome =
  | { kind: 'committed', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'declined' }

/**
 * Commits the changes of the work tree as one operation that Undo can revert
 * (`smith commit <operation>`). It first runs `--dry-run`, shows the files, and
 * only on a yes commits with `--yes`. smith never commits unasked, so this runs
 * only when the person pressed "Commit these changes now".
 * @param deps - how to run smith and ask the person
 * @param operation - the words that name the operation, such as "add app web"
 * @returns how it ended; it never throws
 */
export async function commitChanges (deps: CommitChangesDeps, operation: string): Promise<CommitChangesOutcome> {
  const words = operation.split(' ').filter(Boolean)
  const plan = await deps.run(['commit', ...words, '--dry-run'])
  if (plan.code !== 0) return { kind: 'refused', message: describeFailure(plan) }
  if (!await deps.confirm(plan.stdout.trim())) return { kind: 'declined' }

  const done = await deps.run(['commit', ...words, '--yes'])

  return done.code === 0 ? { kind: 'committed', lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) } : { kind: 'refused', message: describeFailure(done) }
}
