import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import { deriveArguments } from './derive-arguments.mapper.ts'
import type { DeriveAction, DeriveReport } from './derive-report.contract.ts'
import { readDeriveReport } from './derive-report.mapper.ts'

/** Runs smith in the workspace folder with these arguments. */
export type RunSmithIn = (args: string[]) => Promise<SmithRun>

/**
 * Asks smith whether the derived files are as they should be. It writes nothing.
 * @param run - runs smith in the folder
 */
export async function readDeriveStatus (run: RunSmithIn): Promise<DeriveReport> {
  return readDeriveReport(await run(deriveArguments('check')))
}

/**
 * Asks smith who owns a derived file and why it is in its state, with the diff when it drifted.
 * @param run - runs smith in the folder
 * @param file - the file, relative to the folder
 * @returns smith's answer
 * @throws Error with smith's words and the fix, when smith has no answer
 */
export async function explainDerivedFile (run: RunSmithIn, file: string): Promise<string> {
  const result = await run(deriveArguments('explain', file))
  if (result.code !== 0) throw new Error(describeFailure(result))

  return result.stdout
}

/** How a change to a derived file ended: what smith printed, or why it refused. */
export type DeriveChange = { ok: true, output: string } | { ok: false, message: string }

/**
 * Changes derived files: writes them (`apply`), or settles one file that
 * smith holds back (`replace`, `adopt`, `eject`). The caller asks the person first.
 * @param run - runs smith in the folder
 * @param action - what to do
 * @param file - the file, relative to the folder; absent for `apply` of every file
 */
export async function changeDerivedFile (run: RunSmithIn, action: Exclude<DeriveAction, 'check' | 'explain'>, file?: string): Promise<DeriveChange> {
  const result = await run(deriveArguments(action, file))

  return result.code === 0 ? { ok: true, output: result.stdout } : { ok: false, message: describeFailure(result) }
}
