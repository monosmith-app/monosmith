import { describeFailure, type SmithRun } from '../smith-run/index.ts'

/** What the plan of `smith commit --dry-run` says, for the question of which files go in. */
export interface CommitPlan {
  /** The text smith printed: the files of the last operation, and the others git shows. */
  text:   string
  /** The files git shows as changed that the last operation did not change; empty when there are none. */
  extras: string[]
}

/** What committing needs from the editor. */
export interface CommitChangesDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:    (args: string[]) => Promise<SmithRun>
  /** Shows the plan and asks which files go in; undefined means the person cancelled. */
  choose: (plan: CommitPlan) => Promise<CommitScope | undefined>
}

/** Which files go into the commit: those the operation changed, or everything git shows as changed. */
export type CommitScope = 'operation' | 'all'

/** How committing ended: committed (`lines` are what smith printed), refused with smith's words, or declined by the person. */
export type CommitChangesOutcome =
  | { kind: 'committed', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'declined' }

/**
 * Reads the files that git shows and the last operation did not change from the plan.
 * @param text - the output of `smith commit --dry-run`
 * @returns the plan with its extra files
 */
export function readCommitPlan (text: string): CommitPlan {
  const lines = text.split(/\r?\n/)
  const start = lines.findIndex((line) => /Git shows \d+ more/.test(line))
  const after = start === -1 ? [] : lines.slice(start + 1)
  const end = after.findIndex((line) => !/^\s+\S/.test(line))
  const extras = (end === -1 ? after : after.slice(0, end)).map((line) => line.trim())

  return { text: text.trim(), extras }
}

/**
 * Commits the changes of the work tree as one operation that Undo can revert
 * (`smith commit [words]`). It runs `--dry-run`, shows the files, asks which
 * ones go in, and only then commits with `--yes`. smith never commits unasked,
 * so this runs only when the person chose the command or pressed a button.
 * @param deps - how to run smith and ask the person
 * @param operation - the words that name the operation, such as "add app web"; empty for the last operation
 * @returns how it ended; it never throws
 */
export async function commitChanges (deps: CommitChangesDeps, operation = ''): Promise<CommitChangesOutcome> {
  const words = operation.split(' ').filter(Boolean)
  const plan = await deps.run(['commit', ...words, '--dry-run'])
  if (plan.code !== 0) return { kind: 'refused', message: describeFailure(plan) }
  const scope = await deps.choose(readCommitPlan(plan.stdout))
  if (scope === undefined) return { kind: 'declined' }

  const done = await deps.run(['commit', ...words, scope === 'all' ? '--all' : '--only-operation', '--yes'])

  return done.code === 0 ? { kind: 'committed', lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) } : { kind: 'refused', message: describeFailure(done) }
}
