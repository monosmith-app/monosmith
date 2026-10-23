import * as vscode from 'vscode'
import { describeStatus } from './status-text.mapper.ts'

/** The status bar item that shows smith's state. */
export interface SmithStatusBar extends vscode.Disposable {
  update: (version: string | undefined, affected: number | undefined, running: boolean) => void
}

/**
 * @param command - the command a click runs
 */
export function createStatusBar (command: string): SmithStatusBar {
  const item = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 0)
  item.command = command
  const update: SmithStatusBar['update'] = (version, affected, running) => {
    const { text, tooltip } = describeStatus(version, affected, running)
    item.text = text
    item.tooltip = tooltip
  }
  update(undefined, undefined, false)
  item.show()

  return { update, dispose: () => { item.dispose() } }
}
