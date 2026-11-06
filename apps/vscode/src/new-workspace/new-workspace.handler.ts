import * as vscode from 'vscode'
import { chooseFolder, type CommandHost } from '../command-host/index.ts'
import { toNewArguments, type BlankFolder } from '../empty-folder/index.ts'
import { openSmithTerminal, smithForTerminal } from '../smith-terminal/index.ts'
import { askCustomChoices, askSetupWay, askWorkspaceName, confirmNewWorkspace, tellNewWorkspaceOutcome } from './new-workspace-prompts.client.ts'
import { createWorkspace } from './new-workspace.use-case.ts'

/**
 * Creates a workspace in an empty folder with the same two-way start as the TUI
 * (the defaults, or customise). Nothing is written before the person has seen the
 * choices and the list of files and said yes. "Answer in a terminal" keeps smith's own questions.
 * @param host - the folders and the Output channel of the extension
 * @param blanks - the folders without a smith.toml
 * @param created - runs when the workspace exists, so the view reads it
 */
export async function createWorkspaceHere (host: CommandHost, blanks: readonly BlankFolder[], created: () => Promise<void>): Promise<void> {
  const chosen = await chooseFolder(blanks.filter((folder) => folder.kind === 'empty'))
  if (chosen === undefined) return
  const executable = await smithForTerminal(chosen.root)
  if (executable === undefined) return
  host.showOutput()
  const outcome = await createWorkspace({
    folderName: chosen.name,
    run:        host.smithAt(executable, chosen.root, chosen.name, 60_000),
    apply:      async (args) => await vscode.window.withProgress(
      { location: vscode.ProgressLocation.Notification, title: `MonoSmith: creating the workspace in ${chosen.name}…` },
      async () => await host.smithAt(executable, chosen.root, chosen.name, 600_000)(args),
    ),
    askWay:    askSetupWay,
    askName:   askWorkspaceName,
    askCustom: askCustomChoices,
    confirm:   confirmNewWorkspace,
  })
  host.note(`MonoSmith: new workspace in ${chosen.name}: ${outcome.kind === 'refused' ? `stopped: ${outcome.message}` : (outcome.kind === 'terminal' ? 'you chose to answer in a terminal' : outcome.kind)}`)
  if (outcome.kind === 'terminal') {
    host.note(`MonoSmith: creating a new workspace in ${chosen.name} (the folder is empty).`)
    openSmithTerminal(executable, toNewArguments(), chosen.root)

    return
  }
  tellNewWorkspaceOutcome(outcome, host.note)
  if (outcome.kind === 'created') await created()
}
