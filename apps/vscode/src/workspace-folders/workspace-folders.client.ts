import * as vscode from 'vscode'
import type { FolderRef } from './find-smith-roots.policy.ts'

/** The folders the editor has open. */
export function listFolders (): FolderRef[] {
  return (vscode.workspace.workspaceFolders ?? []).map((folder) => ({ root: folder.uri.fsPath, name: folder.name }))
}

/**
 * Listens to folders being added or removed.
 * @returns what stops the listening
 */
export function onFoldersChanged (handler: () => void): vscode.Disposable {
  return vscode.workspace.onDidChangeWorkspaceFolders(() => { handler() })
}

/**
 * Listens to a `smith.toml` appearing or going away in an open folder (a
 * conversion wrote one, or it was deleted).
 * @returns what stops the listening
 */
export function onManifestChanged (handler: () => void): vscode.Disposable {
  const watcher = vscode.workspace.createFileSystemWatcher('**/smith.toml', false, true, false)

  return vscode.Disposable.from(watcher, watcher.onDidCreate(handler), watcher.onDidDelete(handler))
}
