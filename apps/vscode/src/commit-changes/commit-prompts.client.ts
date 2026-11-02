import * as vscode from 'vscode'
import type { CommitChangesOutcome, CommitScope } from './commit-changes.use-case.ts'

/**
 * Asks which files the commit holds. The default, first, leaves the person's own changes out.
 * @param operation - the words that name the operation
 * @returns the scope, or undefined when the person cancels
 */
export async function askCommitScope (operation: string): Promise<CommitScope | undefined> {
  const only = 'Only the files this operation changed'
  const everything = 'Everything git shows as changed'
  const choice = await vscode.window.showInformationMessage(
    `MonoSmith: which files go into the commit "${operation}"?`,
    { modal: true, detail: 'The first choice leaves your own changes in the work tree. The second takes every changed file, your own work included. You see the list before anything is committed.' },
    only,
    everything,
  )

  return choice === only ? 'operation' : (choice === everything ? 'all' : undefined)
}

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
