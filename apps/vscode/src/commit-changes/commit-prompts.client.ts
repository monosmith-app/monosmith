import * as vscode from 'vscode'
import type { CommitChangesOutcome, CommitPlan, CommitScope } from './commit-changes.use-case.ts'

/**
 * Shows the files of the plan and asks which ones go into the commit. With other
 * changes in the work tree the default, first, leaves them out; without them it
 * is a plain yes or no. Nothing is committed before this answer.
 * @param plan - what `smith commit --dry-run` printed, with the extra files
 * @returns the scope, or undefined when the person cancels
 */
export async function chooseCommitScope (plan: CommitPlan): Promise<CommitScope | undefined> {
  if (plan.extras.length === 0) {
    const commit = 'Commit these changes'
    const only = await vscode.window.showInformationMessage(
      'MonoSmith: commit the changes of the last operation?',
      { modal: true, detail: `${plan.text}\n\nIt is one commit that Undo can revert.` },
      commit,
    )

    return only === commit ? 'operation' : undefined
  }
  const only = 'Only the files the last operation changed'
  const everything = `Everything git shows as changed (${plan.extras.length} more)`
  const choice = await vscode.window.showInformationMessage(
    'MonoSmith: which files go into the commit?',
    {
      modal:  true,
      detail: `${plan.text}\n\nThe first choice leaves your own changes in the work tree. The second also takes these files:\n${plan.extras.map((file) => `  ${file}`).join('\n')}\n\nIt is one commit that Undo can revert.`,
    },
    only,
    everything,
  )

  return choice === only ? 'operation' : (choice === everything ? 'all' : undefined)
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
