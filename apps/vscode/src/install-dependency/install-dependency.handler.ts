import * as vscode from 'vscode'
import { askCommitBox } from '../commit-box/index.ts'
import { projectGiven, type CommandHost } from '../command-host/index.ts'
import { installDependency } from './install-dependency.use-case.ts'
import { askInstallRequest, tellDependencyOutcome } from './install-prompts.client.ts'

/**
 * Asks library or registry first, then what to add, and runs `smith install`.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 * @param only - the source the command is limited to, when its name says one
 */
export async function addDependencyHere (host: CommandHost, argument: unknown, only?: 'workspace' | 'registry'): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite('adding a dependency')) return
  const projects = state.snapshot?.projects ?? []
  if (projects.length === 0) {
    void vscode.window.showInformationMessage('MonoSmith: this workspace has no project to add a dependency to.')

    return
  }
  host.showOutput()
  const outcome = await installDependency({
    askCommit: async () => await askCommitBox('add a dependency'),
    ask:       async () => await askInstallRequest(projects.map((project) => ({ name: project.name, kind: project.kind, language: project.language, deps: project.deps })), projectGiven(argument), only),
    apply:     async (args) => await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: `MonoSmith: adding a dependency in ${state.name}…` },
      async () => await host.smithIn(state, true, 600_000)(args),
    ),
  })
  host.report('add dependency', state, outcome)
  tellDependencyOutcome(outcome, host.note)
  if (outcome.kind === 'added') await host.refreshFolder(state.root, 'full', false)
}
