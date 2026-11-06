import type { CommandHost } from '../command-host/index.ts'
import { openSmithTerminal } from '../smith-terminal/index.ts'
import { chooseToolAdmin, chooseToolWay, openToolPage, showToolCommands, tellToolsOutcome } from './install-tools-prompts.client.ts'
import { installTools } from './install-tools.use-case.ts'

/**
 * Offers the ways to get the tools the workspace needs (`smith tools install`): a pick per tool with the command in front of the person.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function installToolsHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined) return
  const outcome = await installTools({
    run:           host.smithIn(state, true, 600_000),
    chooseWay:     chooseToolWay,
    chooseAdmin:   chooseToolAdmin,
    openPage:      openToolPage,
    showCommands:  showToolCommands,
    runInTerminal: (args) => { openSmithTerminal(state.location?.executable ?? 'smith', args, state.root) },
  })
  host.report('tools', state, outcome)
  tellToolsOutcome(outcome)
}
