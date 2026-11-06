import * as vscode from 'vscode'

/** Says updating a dependency needs a trusted workspace. */
export function tellUpdateNeedsTrust (): void {
  void vscode.window.showWarningMessage('MonoSmith: updating dependencies needs a trusted workspace.')
}

/**
 * Asks whether to update a dependency in every project that declares it.
 * @param name - the dependency
 * @returns true on a yes
 */
export async function confirmUpdate (name: string): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    `MonoSmith: update ${name} to its latest version?`,
    { modal: true, detail: 'smith rewrites the manifests of every project that declares it, then installs again. It runs in a terminal so you can read what it does.' },
    'Update',
  )

  return choice === 'Update'
}
