import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import { toTryDocument } from './try-document.mapper.ts'

/** What trying a command needs from the editor. */
export interface TryDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:   (args: string[]) => Promise<SmithRun>
  /** Asks which smith command to try, as the words after `smith`. Resolves undefined when the person cancels. */
  ask:   () => Promise<string[] | undefined>
  /** Shows the page of the result. */
  show:  (page: string) => Promise<void>
  /** Asks whether to run the command for real, after a try that passed. */
  offer: (command: string[]) => Promise<boolean>
}

/** How trying ended: shown (and whether it was then applied), refused with smith's words, or cancelled. */
export type TryOutcome =
  | { kind: 'tried', passed: boolean, applied: boolean } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }

/**
 * Runs a smith command that changes files in a temporary copy
 * (`smith try <command> --json`), shows what it did and which checks passed,
 * and, only after a try that passed and a yes, runs it for real
 * (`smith try <command> --yes`). The try itself leaves the files of the person
 * as they are.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function tryCommand (deps: TryDeps): Promise<TryOutcome> {
  const command = await deps.ask()
  if (command === undefined || command.length === 0) return { kind: 'cancelled' }

  const tried = await deps.run(['try', ...command, '--json'])
  let shown: { page: string, passed: boolean }
  try {
    shown = toTryDocument(tried.stdout)
  } catch {
    return { kind: 'refused', message: describeFailure(tried) }
  }
  await deps.show(shown.page)
  if (!shown.passed || !await deps.offer(command)) return { kind: 'tried', passed: shown.passed, applied: false }

  const done = await deps.run(['try', ...command, '--yes'])

  return done.code === 0 ? { kind: 'tried', passed: true, applied: true } : { kind: 'refused', message: describeFailure(done) }
}
