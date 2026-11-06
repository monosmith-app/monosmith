import * as vscode from 'vscode'
import { parseChoices } from '../add-project/index.ts'
import { projectGiven, type CommandHost } from '../command-host/index.ts'
import { swapChoice } from './swap-choice.use-case.ts'
import { askSwapRequest, confirmSwapPlan, tellSwapOutcome } from './swap-prompts.client.ts'

/**
 * Asks which choice of which project to swap, shows the plan of `smith swap --dry-run`, and swaps on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function swapHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite('swapping a choice')) return
  const given = projectGiven(argument)
  host.showOutput()
  const outcome = await swapChoice({
    run:   host.smithIn(state),
    apply: async (args) => await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: `MonoSmith: swapping in ${state.name}… smith installs and tests the project.` },
      async () => await host.smithIn(state, true, 600_000)(args),
    ),
    parse:   parseChoices,
    ask:     async (choices) => await askSwapRequest(choices, state.snapshot?.projects ?? [], given),
    confirm: confirmSwapPlan,
  })
  host.report('swap', state, outcome)
  tellSwapOutcome(outcome, host.note)
  if (outcome.kind === 'swapped') await host.refreshFolder(state.root, 'full', false)
}
