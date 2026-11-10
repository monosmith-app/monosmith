import { describeFailure, type SmithRun } from '../smith-run/index.ts'

/** Which version wins when projects declare a dependency at different versions: one of `smith sync --prefer`. */
export type SyncRule = 'installed' | 'highest' | 'lowest'

/** What syncing needs from the editor. */
export interface SyncDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:     (args: string[]) => Promise<SmithRun>
  /** Asks which version wins; undefined when the person cancels. */
  choose:  () => Promise<SyncRule | undefined>
  /** Shows the plan smith printed and answers whether to write it. */
  confirm: (plan: string, rule: SyncRule) => Promise<boolean>
}

/** How syncing ended: written (`lines` are what smith printed), refused with smith's words, or declined by the person. */
export type SyncOutcome =
  | { kind: 'synced', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'declined' }

/**
 * Makes every project declare a shared dependency at one version
 * (`smith sync`). The person chooses which version wins, smith prints the plan
 * with `--dry-run`, and only on a yes is it written, with `--yes`.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function syncVersions (deps: SyncDeps): Promise<SyncOutcome> {
  const rule = await deps.choose()
  if (rule === undefined) return { kind: 'declined' }
  const plan = await deps.run(['sync', '--prefer', rule, '--dry-run'])
  if (plan.code !== 0) return { kind: 'refused', message: describeFailure(plan) }
  if (!await deps.confirm(plan.stdout.trim(), rule)) return { kind: 'declined' }

  const done = await deps.run(['sync', '--prefer', rule, '--yes'])

  return done.code === 0 ? { kind: 'synced', lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) } : { kind: 'refused', message: describeFailure(done) }
}
