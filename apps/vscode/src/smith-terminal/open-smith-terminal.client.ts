import * as vscode from 'vscode'
import { toTerminalCommand } from './terminal-command.mapper.ts'

/**
 * Runs smith in a MonoSmith terminal, in the folder, for what asks a person
 * questions or takes long (`smith up`). The terminal stays open so the result
 * can be read.
 * @param executable - the smith the folder runs
 * @param args - plain words, checked by the caller
 * @param cwd - the folder
 */
export function openSmithTerminal (executable: string, args: readonly string[], cwd: string): void {
  const terminal = vscode.window.createTerminal({ name: 'MonoSmith', cwd })
  terminal.show()
  terminal.sendText(toTerminalCommand(executable, args, vscode.env.shell))
}
