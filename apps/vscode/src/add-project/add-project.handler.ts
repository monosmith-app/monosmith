import * as vscode from 'vscode'
import { commitAfterTheFact } from '../commit-changes/index.ts'
import type { CommandHost } from '../command-host/index.ts'
import { parseWorkspaceEnvironments } from '../guided-commands/index.ts'
import { openSmithTerminal } from '../smith-terminal/index.ts'
import { addProject } from './add-project.use-case.ts'
import { askAddRequest, confirmAddPlan, tellAddOutcome } from './add-prompts.client.ts'
import { checkPicks } from './check-picks.use-case.ts'
import { outcomeWords } from './outcome-words.mapper.ts'

/**
 * Asks what to add, shows the plan of `smith add --dry-run`, and adds on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function addProjectHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite('adding a project')) return
  host.showOutput()
  const outcome = await addProject({
    run:   host.smithIn(state),
    apply: async (args) => await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: `MonoSmith: adding to ${state.name}… smith installs and tests the new project.` },
      async () => await host.smithIn(state, true, 600_000)(args),
    ),
    ask: async (options) => {
      const configured = await host.smithIn(state)(['config', 'get', 'workspace.environments'])

      return await askAddRequest(options, state.snapshot?.projects.map((project) => project.name) ?? [], state.snapshot?.projects ?? [], configured.code === 0 ? parseWorkspaceEnvironments(configured.stdout) : [], async (asked) => await checkPicks(host.smithIn(state), options, asked))
    },
    confirm: confirmAddPlan,
  })
  host.note(`MonoSmith: add in ${state.name}: ${outcomeWords(outcome)}`)
  if (outcome.kind === 'composing') {
    await vscode.commands.executeCommand('monosmith.composeStack', argument)

    return
  }
  tellAddOutcome(outcome, host.note, {
    commit:     commitAfterTheFact(host, state),
    checkAgain: (args) => { openSmithTerminal(state.location?.executable ?? 'smith', args, state.root) },
  })
  if (['added', 'unchecked', 'failing'].includes(outcome.kind)) await host.refreshFolder(state.root, 'full', false)
}
