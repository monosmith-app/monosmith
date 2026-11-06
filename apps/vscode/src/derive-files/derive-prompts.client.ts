import * as vscode from 'vscode'
import type { DeriveAction } from './derive-report.contract.ts'
import { say } from '../words/index.ts'

const WORDS: Record<'apply' | 'replace' | 'adopt' | 'eject', { button: string, effect: string }> = {
  apply:   { button: say('vscode.derive-files.apply'), effect: say('vscode.derive-files.smith-writes-every-derived-file-that') },
  replace: { button: say('vscode.derive-files.overwrite'), effect: say('vscode.derive-files.smith-overwrites-the-file-with-the') },
  adopt:   { button: say('vscode.derive-files.adopt'), effect: say('vscode.derive-files.smith-keeps-your-edit-as-the') },
  eject:   { button: say('vscode.derive-files.eject'), effect: say('vscode.derive-files.the-file-becomes-yours-smith-writes') },
}

/**
 * Asks before a change to a derived file, and says what the change does.
 * @param action - what the person chose
 * @param file - the file, or undefined for every derived file
 * @returns `yes`, `yes-commit` (make it one commit, so Undo works), `explain` (show the diff of the file first), `preview` (show what a plain derive would write) or `no`
 */
export async function confirmDeriveChange (action: 'apply' | 'replace' | 'adopt' | 'eject', file: string | undefined): Promise<'yes' | 'yes-commit' | 'explain' | 'preview' | 'no'> {
  const words = WORDS[action]
  const explain = say('vscode.derive-files.explain-first')
  const preview = say('vscode.derive-files.preview')
  const commit = say('vscode.derive-files.and-commit-so-undo-works', { button: words.button })
  const choice = await vscode.window.showWarningMessage(
    say('vscode.derive-files.monosmith', { toLowerCase: words.button.toLowerCase(), part: file ?? say('vscode.derive-files.the-derived-files') }),
    { modal: true, detail: say('vscode.derive-files.by-default-smith-leaves-the-files', { effect: words.effect }) },
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
    void vscode.window.showInformationMessage(say('vscode.derive-files.monosmith-commit-it-together-with-smith', { part: file ?? say('vscode.derive-files.the-derived-files'), part2: action === 'eject' ? say('vscode.derive-files.is-now-yours') : (say(action === 'adopt' ? 'vscode.derive-files.keeps-your-edit' : 'vscode.derive-files.is-written')) }))
  } else {
    void vscode.window.showWarningMessage(say('vscode.convert-workspace.monosmith-nothing-was-changed', { message }))
  }
}

/** Says a derived file must be chosen first. */
export function tellChooseDerivedFile (): void {
  void vscode.window.showInformationMessage(say('vscode.derive-files.monosmith-choose-a-derived-file-first'))
}

/**
 * Says smith has nothing to write for a file.
 * @param file - the file, relative to its folder
 */
export function tellNothingToWrite (file: string): void {
  void vscode.window.showInformationMessage(say('vscode.derive-files.monosmith-smith-has-nothing-to-write', { file }))
}
