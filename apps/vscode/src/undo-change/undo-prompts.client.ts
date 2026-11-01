import * as vscode from 'vscode'
import type { UndoOutcome } from './undo-change.use-case.ts'

/**
 * Shows what undoing would revert and asks.
 * @param plan - the text of `smith undo --dry-run`
 * @param folder - the name of the workspace folder
 * @returns whether the person agreed
 */
export async function confirmUndo (plan: string, folder: string): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    `MonoSmith: undo the last change in ${folder}?`,
    { modal: true, detail: `${plan}\n\nsmith reverts it with git revert: a new commit, and your history stays.` },
    'Undo',
  )

  return choice === 'Undo'
}

/**
 * Says how undoing ended. It does not wait for the person to read it.
 * @param outcome - what `undoLastChange` returned
 */
export function tellUndoOutcome (outcome: UndoOutcome): void {
  switch (outcome.kind) {
    case 'undone': {
      void vscode.window.showInformationMessage('MonoSmith: the last change is reverted in a new commit.')
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was undone. ${outcome.message}`)
      break
    }
    case 'declined': {
      break
    }
  }
}
