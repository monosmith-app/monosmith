import { tmpdir } from 'node:os'
import * as vscode from 'vscode'
import type { CommandHost } from '../command-host/index.ts'
import { reportFile, writeReport } from './write-report.use-case.ts'

/**
 * Writes the HTML report of the workspace to a temp file and opens it in the browser.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function openReport (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined) return
  const outcome = await vscode.window.withProgress(
    { location: vscode.ProgressLocation.Notification, title: `MonoSmith: writing the report of ${state.name}…` },
    async () => await writeReport(host.smithIn(state), reportFile(tmpdir(), state.name)),
  )
  if (!outcome.ok) {
    void vscode.window.showWarningMessage(`MonoSmith: the report was not written. ${outcome.message}`)

    return
  }
  host.note(`MonoSmith: report of ${state.name}: ${outcome.file}`)
  await vscode.env.openExternal(vscode.Uri.file(outcome.file))
}
