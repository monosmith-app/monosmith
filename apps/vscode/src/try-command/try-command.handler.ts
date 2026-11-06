import type { CommandHost } from '../command-host/index.ts'
import { tryCommand } from './try-command.use-case.ts'
import { askTryCommand, offerTryApply, tellTryOutcome } from './try-prompts.client.ts'

/**
 * Runs a smith command in a temporary copy (`smith try`), shows what it did, and runs it for real on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param showPage - shows a page of text in the editor under a name
 * @param argument - what the editor passed to the command
 */
export async function tryHere (host: CommandHost, showPage: (name: string, text: string) => Promise<void>, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite('running a command for real')) return
  const outcome = await tryCommand({
    run:   host.smithIn(state),
    ask:   askTryCommand,
    show:  async (page) => { await showPage('try.md', page) },
    offer: offerTryApply,
  })
  host.report('try', state, outcome)
  tellTryOutcome(outcome)
  if (outcome.kind === 'tried' && outcome.applied) await host.refreshFolder(state.root, 'full', false)
}
