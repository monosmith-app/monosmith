import type { CommandHost } from '../command-host/index.ts'
import { undoLastChange } from './undo-change.use-case.ts'
import { confirmUndo, tellUndoOutcome } from './undo-prompts.client.ts'
import { say } from '../words/index.ts'

/**
 * Shows what the last change smith committed did, and reverts it on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function undoHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite(say('vscode.undo-change.undoing-a-change'))) return
  const outcome = await undoLastChange({ run: host.smithIn(state), confirm: async (plan) => await confirmUndo(plan, state.name) })
  host.report('undo', state, outcome)
  tellUndoOutcome(outcome)
  if (outcome.kind === 'undone') await host.refreshFolder(state.root, 'full', false)
}
