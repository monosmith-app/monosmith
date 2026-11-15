import * as vscode from 'vscode'
import { statusMenuEntries, type MenuState } from './status-menu.mapper.ts'
import { say } from '../words/index.ts'

/** Shows the menu of the status bar item and runs the command that was picked. */
export async function showStatusMenu (state: MenuState): Promise<void> {
  const entries = statusMenuEntries(state)
  const picked = await vscode.window.showQuickPick(
    entries.map((entry) => ({ label: entry.label, description: entry.description, command: entry.command })),
    { title: say('vscode.package.display-name'), placeHolder: say('vscode.quick-pick.what-next') },
  )
  if (picked !== undefined) await vscode.commands.executeCommand(picked.command)
}
