import { delimiter } from 'node:path'
import * as vscode from 'vscode'
import { terminalLaunch } from './terminal-command.mapper.ts'

/**
 * Runs smith in a MonoSmith terminal, in the folder, for what asks a person
 * questions or takes long (`smith up`). The terminal stays open so the result
 * can be read. The line is `smith <arguments>`, as the person could type it:
 * the folder of the smith the workspace runs is put first on the PATH of the terminal.
 * @param executable - the smith the folder runs
 * @param args - plain words, checked by the caller
 * @param cwd - the folder
 * @returns the terminal, so a caller can watch the command end
 */
export function openSmithTerminal (executable: string, args: readonly string[], cwd: string): vscode.Terminal {
  const { line, directory } = terminalLaunch(executable, args, vscode.env.shell)
  const pathKey = Object.keys(process.env).find((key) => key.toLowerCase() === 'path') ?? 'PATH'
  const terminal = vscode.window.createTerminal({
    name: 'MonoSmith',
    cwd,
    ...(directory !== undefined && { env: { [pathKey]: `${directory}${delimiter}${process.env[pathKey] ?? ''}` } }),
  })
  terminal.show()
  terminal.sendText(line)

  return terminal
}
