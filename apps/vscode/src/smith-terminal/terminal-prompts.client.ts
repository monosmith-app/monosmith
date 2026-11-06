import * as vscode from 'vscode'

const INSTALL_GLOBAL_COMMAND = 'npm i -g monosmith'

/** Says smith is not installed on this computer, and offers to copy the command that installs it. */
export async function tellSmithNotInstalled (): Promise<void> {
  const copy = `Copy "${INSTALL_GLOBAL_COMMAND}"`
  const choice = await vscode.window.showWarningMessage('MonoSmith: smith is not installed on this computer. Install it, then try again.', copy)
  if (choice === copy) await vscode.env.clipboard.writeText(INSTALL_GLOBAL_COMMAND)
}
