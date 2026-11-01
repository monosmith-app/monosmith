import * as vscode from 'vscode'
import { toTerminalCommand } from './terminal-command.mapper.ts'

/**
 * Runs smith in a MonoSmith terminal, in the folder, for what asks a person
 * questions or takes long (`smith up`). The terminal stays open so the result
 * can be read.
 * @param executable - the smith the folder runs
 * @param args - plain words, checked by the caller
 * @param cwd - the folder
 * @returns the terminal, so a caller can watch the command end
 */
export function openSmithTerminal (executable: string, args: readonly string[], cwd: string): vscode.Terminal {
  const terminal = vscode.window.createTerminal({ name: 'MonoSmith', cwd })
  terminal.show()
  terminal.sendText(toTerminalCommand(executable, args, vscode.env.shell))

  return terminal
}
