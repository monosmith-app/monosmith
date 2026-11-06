import * as vscode from 'vscode'
import { say } from '../words/index.ts'

const INSTALL_GLOBAL_COMMAND = 'npm i -g monosmith'

/** Says smith is not installed on this computer, and offers to copy the command that installs it. */
export async function tellSmithNotInstalled (): Promise<void> {
  const copy = say('vscode.smith-terminal.copy', { installGlobalCommand: INSTALL_GLOBAL_COMMAND })
  const choice = await vscode.window.showWarningMessage(say('vscode.smith-terminal.monosmith-smith-is-not-installed-on'), copy)
  if (choice === copy) await vscode.env.clipboard.writeText(INSTALL_GLOBAL_COMMAND)
}
