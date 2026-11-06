import * as vscode from 'vscode'
import { outcomeWords } from '../add-project/index.ts'
import { commitAfterTheFact } from '../commit-changes/index.ts'
import type { CommandHost } from '../command-host/index.ts'
import { answerQuestion, askProjectName, confirmComposePlan, pickItems, tellComposeOutcome } from './compose-prompts.client.ts'
import { composeStack } from './compose-stack.use-case.ts'
import { say } from '../words/index.ts'

/**
 * Searches the catalogue, composes the stack, shows the plan of `smith add`, and adds on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function composeStackHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite(say('vscode.compose-stack.composing-a-stack'))) return
  host.showOutput()
  const existing = state.snapshot?.projects.map((project) => project.name) ?? []
  const outcome = await composeStack({
    run:   host.smithIn(state),
    apply: async (args) => await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: say('vscode.compose-stack.monosmith-composing-in-smith-installs-and', { name: state.name }) },
      async () => await host.smithIn(state, true, 600_000)(args),
    ),
    pick:         pickItems,
    answer:       answerQuestion,
    askName:      async (stack) => await askProjectName(existing, stack),
    tell:         (message) => { void vscode.window.showInformationMessage(say('vscode.compose-stack.monosmith', { message })) },
    confirm:      confirmComposePlan,
    installTools: async () => { await vscode.commands.executeCommand('monosmith.installTools', argument) },
  })
  host.note(say('vscode.compose-stack.monosmith-compose-in', { name: state.name, outcomeWords: outcomeWords(outcome) }))
  tellComposeOutcome(outcome, host.note, { commit: commitAfterTheFact(host, state) })
  if (outcome.kind === 'added' || outcome.kind === 'written') await host.refreshFolder(state.root, 'full', false)
}
