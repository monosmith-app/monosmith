import * as vscode from 'vscode'
import type { CommitChangesOutcome } from './commit-changes.use-case.ts'

/**
 * Shows the files that would be committed and asks.
 * @param plan - the text of `smith commit --dry-run`
 * @param operation - the words that name the operation
 * @returns whether the person agreed
 */
export async function confirmCommit (plan: string, operation: string): Promise<boolean> {
  const choice = await vscode.window.showInformationMessage(
    `MonoSmith: commit these changes as "${operation}"?`,
    { modal: true, detail: `${plan}\n\nThe list is everything git shows as changed: check it. It is one commit that Undo can revert.` },
    'Commit these changes',
  )

  return choice === 'Commit these changes'
}

/**
 * Says how committing ended. It does not wait for the person to read it.
 * @param outcome - what `commitChanges` returned
 */
export function tellCommitOutcome (outcome: CommitChangesOutcome): void {
  switch (outcome.kind) {
    case 'committed': {
      void vscode.window.showInformationMessage(`MonoSmith: ${outcome.lines.at(-1) ?? 'the changes are committed.'}`)
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was committed. ${outcome.message}`)
      break
    }
    case 'declined': {
      break
    }
  }
}
