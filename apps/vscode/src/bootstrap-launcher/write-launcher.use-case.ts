import { describeFailure, type SmithRun } from '../smith-run/index.ts'

/** What writing the launcher needs from the editor. */
export interface LauncherDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:     (args: string[]) => Promise<SmithRun>
  /** Asks whether to write the launcher files. */
  confirm: () => Promise<boolean>
}

/** How writing ended: written (`lines` are what smith printed), refused with smith's words, or declined by the person. */
export type LauncherOutcome =
  | { kind: 'written', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'declined' }

/**
 * Writes the launcher that makes a laptop and CI run the same smith
 * (`smith bootstrap`: tools/smith/smith and tools/smith/smith.cmd). It asks
 * first, because the files are meant to be committed.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function writeLauncher (deps: LauncherDeps): Promise<LauncherOutcome> {
  if (!await deps.confirm()) return { kind: 'declined' }
  const done = await deps.run(['bootstrap'])

  return done.code === 0 ? { kind: 'written', lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) } : { kind: 'refused', message: describeFailure(done) }
}
