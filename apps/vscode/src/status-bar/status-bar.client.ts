import * as vscode from 'vscode'
import { describeStatus, type StatusInput } from './status-text.mapper.ts'

/** The status bar item that shows smith's state. */
export interface SmithStatusBar extends vscode.Disposable {
  update: (input: StatusInput) => void
}

/**
 * @param command - the command a click runs
 */
export function createStatusBar (command: string): SmithStatusBar {
  const item = vscode.window.createStatusBarItem('monosmith.status', vscode.StatusBarAlignment.Left, 0)
  item.name = 'MonoSmith'
  item.command = command
  item.accessibilityInformation = { label: 'MonoSmith status, open menu', role: 'button' }
  const update: SmithStatusBar['update'] = (input) => {
    const { text, tooltip, warning } = describeStatus(input)
    item.text = text
    item.tooltip = tooltip
    item.backgroundColor = warning ? new vscode.ThemeColor('statusBarItem.warningBackground') : undefined
    item.accessibilityInformation = { label: `MonoSmith: ${text.replaceAll(/\$\([^)]*\)\s*/g, '')}. Open menu.`, role: 'button' }
  }
  update({ state: 'starting', running: false, failed: 0, origin: [] })
  item.show()

  return { update, dispose: () => { item.dispose() } }
}
