import type { CommandHost } from '../command-host/index.ts'
import type { FolderState } from '../folder-session/index.ts'
import { commitChanges } from './commit-changes.use-case.ts'
import { chooseCommitScope, tellCommitOutcome } from './commit-prompts.client.ts'

/**
 * Commits the changes of an operation after the fact: shows the files, asks which ones go in, and commits on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param state - the folder with the changes
 * @returns a function that takes the name of the operation
 */
export function commitAfterTheFact (host: CommandHost, state: FolderState): (operation: string) => Promise<void> {
  return async (operation) => {
    const done = await commitChanges({ run: host.smithIn(state), choose: chooseCommitScope }, operation)
    host.report('commit', state, done)
    tellCommitOutcome(done)
  }
}

/**
 * Commits the changes of the last operation: shows the files, asks which ones go in, and commits on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function commitHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite('committing')) return
  host.showOutput()
  await commitAfterTheFact(host, state)('')
}
