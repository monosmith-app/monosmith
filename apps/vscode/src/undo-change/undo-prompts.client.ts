import * as vscode from 'vscode'
import type { UndoOutcome } from './undo-change.use-case.ts'
import { say } from '../words/index.ts'

/**
 * Shows what undoing would revert and asks.
 * @param plan - the text of `smith undo --dry-run`
 * @param folder - the name of the workspace folder
 * @returns whether the person agreed
 */
export async function confirmUndo (plan: string, folder: string): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    say('vscode.undo-change.monosmith-undo-the-last-change-in', { folder }),
    { modal: true, detail: say('vscode.undo-change.smith-reverts-it-with-git-revert', { plan }) },
    say('vscode.undo-change.undo'),
  )

  return choice === say('vscode.undo-change.undo')
}

/**
 * Says how undoing ended. It does not wait for the person to read it.
 * @param outcome - what `undoLastChange` returned
 */
export function tellUndoOutcome (outcome: UndoOutcome): void {
  switch (outcome.kind) {
    case 'undone': {
      void vscode.window.showInformationMessage(say('vscode.undo-change.monosmith-the-last-change-is-reverted'))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.undo-change.monosmith-nothing-was-undone', { message: outcome.message }))
      break
    }
    case 'declined': {
      break
    }
  }
}
