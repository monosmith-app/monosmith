import { existsSync } from 'node:fs'
import * as vscode from 'vscode'
import { createStatusBar } from './status-bar/index.ts'
import { locateSmith, startSmith, type SmithSession } from './smith-client/index.ts'
import { cancelRun, createTaskStatusStore, describeEvent, runTask } from './task-run/index.ts'
import { createWorkspaceTree, type TreeNode } from './workspace-view/index.ts'
import { readWorkspace } from './workspace/index.ts'

const INSTALL_COMMAND = 'npm i -D monosmith'

let current: SmithSession | undefined

/**
 * Starts the extension: the MonoSmith view, its commands and the status bar,
 * all served by the smith the team has installed.
 * @param context - the extension's lifetime
 */
export function activate (context: vscode.ExtensionContext): void {
  const folder = vscode.workspace.workspaceFolders?.[0]
  if (folder === undefined) return
  const root = folder.uri.fsPath
  const output = vscode.window.createOutputChannel('MonoSmith')
  const statuses = createTaskStatusStore()
  const tree = createWorkspaceTree(statuses.statusOf)
  const bar = createStatusBar('monosmith.refresh')
  let affected: number | undefined
  let activeRun: string | undefined

  const paint = (): void => { bar.update(current?.version, affected, activeRun !== undefined) }
  const fail = (error: unknown): void => { void vscode.window.showErrorMessage(`MonoSmith: ${error instanceof Error ? error.message : String(error)}`) }

  const connect = async (): Promise<SmithSession | undefined> => {
    if (current !== undefined) return current
    const executable = locateSmith({
      setting:       vscode.workspace.getConfiguration('monosmith').get<string>('path', ''),
      workspaceRoot: root,
      platform:      process.platform,
      pathVariable:  process.env.PATH ?? '',
      exists:        existsSync,
    })
    if (executable === undefined) {
      void offerInstall()

      return undefined
    }
    try {
      current = await startSmith(executable, root)
    } catch (error) {
      fail(error)

      return undefined
    }
    if (current.protocol === 'newer') output.appendLine('smith speaks a newer protocol than this extension knows; it keeps working with what it knows.')

    return current
  }

  const refresh = async (): Promise<void> => {
    const session = await connect()
    if (session !== undefined) {
      try {
        const snapshot = await readWorkspace(session.connection)
        affected = snapshot.affected.size
        tree.show(snapshot)
      } catch (error) {
        session.stop()
        current = undefined
        fail(error)
      }
    }
    paint()
  }

  const run = async (node?: TreeNode): Promise<void> => {
    if (node?.project === undefined || node.task === undefined) return
    if (activeRun !== undefined) {
      void vscode.window.showWarningMessage('MonoSmith: a run is already in progress.')

      return
    }
    const session = await connect()
    if (session === undefined) return
    const runId = `ide-${Date.now()}`
    activeRun = runId
    paint()
    output.show(true)
    try {
      const result = await runTask(session.connection, { runId, task: node.task, project: node.project }, (event) => {
        statuses.apply(event)
        const line = describeEvent(event)
        if (line !== undefined) output.appendLine(line)
      })
      if (!result.ok) void vscode.window.showWarningMessage(`MonoSmith: ${node.project}:${node.task} did not pass. See the output.`)
    } catch (error) {
      output.appendLine(`MonoSmith: ${error instanceof Error ? error.message : String(error)}`)
    } finally {
      activeRun = undefined
      paint()
    }
  }

  const cancel = async (): Promise<void> => {
    if (current === undefined || activeRun === undefined) return
    await cancelRun(current.connection, activeRun)
  }

  context.subscriptions.push(
    output,
    bar,
    { dispose: statuses.onChange(tree.redraw) },
    vscode.window.registerTreeDataProvider('monosmith.projects', tree.provider),
    vscode.commands.registerCommand('monosmith.refresh', refresh),
    vscode.commands.registerCommand('monosmith.runTask', run),
    vscode.commands.registerCommand('monosmith.cancelRun', cancel),
  )
  void refresh()
}

/** Ends `smith serve` with the extension. */
export function deactivate (): void {
  current?.stop()
  current = undefined
}

async function offerInstall (): Promise<void> {
  const copy = `Copy "${INSTALL_COMMAND}"`
  const choice = await vscode.window.showErrorMessage(
    'MonoSmith: smith was not found. Install it in the workspace, put it on the PATH, or set monosmith.path.',
    copy,
    'Open settings',
  )
  if (choice === copy) await vscode.env.clipboard.writeText(INSTALL_COMMAND)
  else if (choice === 'Open settings') await vscode.commands.executeCommand('workbench.action.openSettings', 'monosmith.path')
}
