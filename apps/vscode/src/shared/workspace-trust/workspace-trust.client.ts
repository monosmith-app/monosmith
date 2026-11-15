import * as vscode from 'vscode'

/** Whether the user trusts the folders that are open. */
export function isWorkspaceTrusted (): boolean {
  return vscode.workspace.isTrusted
}

/**
 * Listens to the user granting trust to the workspace.
 * @returns what stops the listening
 */
export function onTrustGranted (handler: () => void): vscode.Disposable {
  return vscode.workspace.onDidGrantWorkspaceTrust(() => { handler() })
}
