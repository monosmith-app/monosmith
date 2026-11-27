import * as vscode from 'vscode'
import { say } from '../../shared/words/index.ts'

/** Says updating a dependency needs a trusted workspace. */
export function tellUpdateNeedsTrust (): void {
  void vscode.window.showWarningMessage(say('vscode.dependencies.monosmith-updating-dependencies-needs-a-trusted'))
}

/** Says no update is a patch or minor release smith can rewrite. */
export function tellNothingSafe (): void {
  void vscode.window.showInformationMessage(say('vscode.dependencies.nothing-safe'))
}

/**
 * Asks whether to update dependencies in every project that declares them, with the command it runs.
 * @param names - the dependencies
 * @param line - the smith command line the update runs
 * @returns true on a yes
 */
export async function confirmUpdate (names: readonly string[], line: string): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    say('vscode.dependencies.monosmith-update-to-its-latest-version', { name: names.join(', ') }),
    { modal: true, detail: `$ ${line}\n\n${say('vscode.dependencies.smith-rewrites-the-manifests-of-every')}` },
    say('vscode.dependencies.update'),
  )

  return choice === say('vscode.dependencies.update')
}

/**
 * Asks whether to open the dependency repair in a terminal, where smith asks what to repair.
 * @param line - the smith command line it opens
 * @returns true on a yes
 */
export async function confirmRepair (line: string): Promise<boolean> {
  const open = say('vscode.dependencies.open-repair')
  const choice = await vscode.window.showWarningMessage(say('vscode.dependencies.repair-question'), { modal: true, detail: `$ ${line}\n\n${say('vscode.dependencies.repair-detail')}` }, open)

  return choice === open
}
