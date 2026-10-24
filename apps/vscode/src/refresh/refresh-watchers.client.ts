import * as vscode from 'vscode'
import type { RefreshScope } from './refresh-scope.contract.ts'
import { isRelevantSave } from './save-relevance.policy.ts'

/**
 * Watches one folder for what makes the view stale: `smith.toml` edited
 * (the projects may differ), the git HEAD moved (a branch switch), and a file
 * saved (the affected projects may differ). Every event is only a request; the
 * caller coalesces them.
 * @param root - the folder to watch
 * @param request - asks for a refresh of the given scope
 */
export function watchRefreshTriggers (root: string, request: (scope: RefreshScope) => void): vscode.Disposable {
  const base = vscode.Uri.file(root)
  const toml = vscode.workspace.createFileSystemWatcher(new vscode.RelativePattern(base, '**/smith.toml'))
  const head = vscode.workspace.createFileSystemWatcher(new vscode.RelativePattern(base, '.git/HEAD'))
  const fullFor = (uri: vscode.Uri): void => {
    if (isRelevantSave(root, uri.fsPath)) request('full')
  }
  const subscriptions = [
    toml,
    head,
    toml.onDidChange(fullFor),
    toml.onDidCreate(fullFor),
    toml.onDidDelete(fullFor),
    head.onDidChange(() => { request('full') }),
    head.onDidCreate(() => { request('full') }),
    vscode.workspace.onDidSaveTextDocument((document) => {
      if (document.uri.scheme === 'file' && isRelevantSave(root, document.uri.fsPath)) request('affected')
    }),
  ]

  return vscode.Disposable.from(...subscriptions)
}
