import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import { installArguments } from './install-arguments.mapper.ts'
import type { InstallOutcome, InstallRequest } from './install-request.contract.ts'

/** What adding a dependency needs from the editor. */
export interface InstallDeps {
  /** Asks the questions: library or registry first, then the library and range, or the packages. Resolves undefined when the person cancels. */
  ask:   () => Promise<InstallRequest | undefined>
  /** Runs smith in the workspace folder to its end; the package manager may take long. */
  apply: (args: string[]) => Promise<SmithRun>
}

/**
 * Adds a dependency to a project: asks the person, then runs `smith install`. smith puts the manifest back
 * when the install fails, so a refused request leaves the project as it was.
 * @param deps - how to ask and how to run smith
 * @returns how it ended; it never throws
 */
export async function installDependency (deps: InstallDeps): Promise<InstallOutcome> {
  const request = await deps.ask()
  if (request === undefined) return { kind: 'cancelled' }
  const done = await deps.apply(installArguments(request))

  return done.code === 0 ? { kind: 'added', lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) } : { kind: 'refused', message: describeFailure(done) }
}
