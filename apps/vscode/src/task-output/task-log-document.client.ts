import * as vscode from 'vscode'
import type { TaskLogStore } from './task-log.store.ts'

/** The scheme the logs of tasks are opened under. */
export const LOG_SCHEME = 'monosmith-log'

/** Opens the last log of a task as a read-only document. */
export interface TaskLogDocuments extends vscode.Disposable {
  /**
   * Opens the log of a task beside the active editor.
   * @param root - the folder the task belongs to
   * @param task - the task id (`web:build`)
   */
  open:    (root: string, task: string) => Promise<void>
  /** Says a log has new lines, so an open document shows them. */
  changed: (root: string, task: string) => void
  /** Which task a log document is of. */
  taskOf:  (uri: vscode.Uri) => { root: string, task: string } | undefined
}

/** The store key of a task of a folder. */
export function logKey (root: string, task: string): string {
  return `${root}\0${task}`
}

/** Registers the document scheme the logs are served under. */
export function createTaskLogDocuments (logs: TaskLogStore): TaskLogDocuments {
  const emitter = new vscode.EventEmitter<vscode.Uri>()
  const uriOf = (root: string, task: string): vscode.Uri => vscode.Uri.from({
    scheme: LOG_SCHEME,
    path:   `/${task.replaceAll(':', '-')}.log`,
    query:  new URLSearchParams({ root, task }).toString(),
  })
  const taskOf: TaskLogDocuments['taskOf'] = (uri) => {
    if (uri.scheme !== LOG_SCHEME) return
    const query = new URLSearchParams(uri.query)
    const root = query.get('root')
    const task = query.get('task')

    return root === null || task === null ? undefined : { root, task }
  }
  const registration = vscode.workspace.registerTextDocumentContentProvider(LOG_SCHEME, {
    onDidChange:                emitter.event,
    provideTextDocumentContent: (uri) => {
      const which = taskOf(uri)
      const lines = which === undefined ? undefined : logs.lines(logKey(which.root, which.task))

      return lines === undefined ? 'No output was kept for this task. Run it to see its output here.' : lines.join('\n')
    },
  })

  return {
    open: async (root, task) => {
      const document = await vscode.workspace.openTextDocument(uriOf(root, task))
      await vscode.window.showTextDocument(document, { preview: true, preserveFocus: false })
    },
    changed: (root, task) => { emitter.fire(uriOf(root, task)) },
    taskOf,
    dispose: () => { registration.dispose(); emitter.dispose() },
  }
}
