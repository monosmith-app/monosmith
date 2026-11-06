import * as vscode from 'vscode'
import { say } from '../words/index.ts'

/** Says updating a dependency needs a trusted workspace. */
export function tellUpdateNeedsTrust (): void {
  void vscode.window.showWarningMessage(say('vscode.dependencies.monosmith-updating-dependencies-needs-a-trusted'))
}

/**
 * Asks whether to update a dependency in every project that declares it.
 * @param name - the dependency
 * @returns true on a yes
 */
export async function confirmUpdate (name: string): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    say('vscode.dependencies.monosmith-update-to-its-latest-version', { name }),
    { modal: true, detail: say('vscode.dependencies.smith-rewrites-the-manifests-of-every') },
    say('vscode.dependencies.update'),
  )

  return choice === say('vscode.dependencies.update')
}
