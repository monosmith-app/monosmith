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
