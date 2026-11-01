import * as vscode from 'vscode'
import type { DeriveAction } from './derive-report.contract.ts'

const WORDS: Record<'apply' | 'replace' | 'adopt' | 'eject', { button: string, effect: string }> = {
  apply:   { button: 'Apply', effect: 'smith writes every derived file that is stale or missing and records it in smith.lock. It leaves the files someone edited alone.' },
  replace: { button: 'Overwrite', effect: 'smith overwrites the file with the derived content. Your edit is lost; git can bring it back.' },
  adopt:   { button: 'Adopt', effect: 'smith keeps your edit as the new baseline. This works for a file smith owns by keys that still has every derived key. If the inputs change later, the file is in drift again.' },
  eject:   { button: 'Eject', effect: 'The file becomes yours. smith writes "you" as its owner in smith.lock and never touches it again.' },
}

/**
 * Asks before a change to a derived file, and says what the change does.
 * @param action - what the person chose
 * @param file - the file, or undefined for every derived file
 * @returns `yes`, `yes-commit` (make it one commit, so Undo works), `explain` (show the diff of the file first), `preview` (show what a plain derive would write) or `no`
 */
export async function confirmDeriveChange (action: 'apply' | 'replace' | 'adopt' | 'eject', file: string | undefined): Promise<'yes' | 'yes-commit' | 'explain' | 'preview' | 'no'> {
  const words = WORDS[action]
  const explain = 'Explain first'
  const preview = 'Preview'
  const commit = `${words.button} and commit (so Undo works)`
  const choice = await vscode.window.showWarningMessage(
    `MonoSmith: ${words.button.toLowerCase()} ${file ?? 'the derived files'}?`,
    { modal: true, detail: `${words.effect}\n\nBy default smith leaves the files uncommitted: review them with git. The commit button makes it one commit that Undo can revert.` },
    words.button,
    commit,
    file === undefined ? preview : explain,
  )
  if (choice === words.button) return 'yes'
  if (choice === commit) return 'yes-commit'
  if (choice === preview) return 'preview'

  return choice === explain ? 'explain' : 'no'
}

/**
 * Says how a change to derived files ended. It does not wait for the person to read it.
 * @param action - what was done
 * @param file - the file, or undefined for every derived file
 * @param ok - whether smith did it
 * @param message - smith's words when it did not, with the fix
 */
export function tellDeriveChange (action: Exclude<DeriveAction, 'check' | 'explain'>, file: string | undefined, ok: boolean, message: string): void {
  if (ok) {
    void vscode.window.showInformationMessage(`MonoSmith: ${file ?? 'the derived files'} ${action === 'eject' ? 'is now yours' : (action === 'adopt' ? 'keeps your edit' : 'is written')}. Commit it together with smith.lock.`)
  } else {
    void vscode.window.showWarningMessage(`MonoSmith: nothing was changed. ${message}`)
  }
}
