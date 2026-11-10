import * as vscode from 'vscode'
import type { CommitChangesOutcome, CommitPlan, CommitScope } from './commit-changes.use-case.ts'
import { say } from '../words/index.ts'

/**
 * Shows the files of the plan and asks which ones go into the commit. With other
 * changes in the work tree the default, first, leaves them out; without them it
 * is a plain yes or no. Nothing is committed before this answer.
 * @param plan - what `smith commit --dry-run` printed, with the extra files
 * @returns the scope, or undefined when the person cancels
 */
export async function chooseCommitScope (plan: CommitPlan): Promise<CommitScope | undefined> {
  if (plan.extras.length === 0) {
    const commit = say('commands.describe.commit.title')
    const only = await vscode.window.showInformationMessage(
      say('vscode.commit-changes.monosmith-commit-the-changes-of-the'),
      { modal: true, detail: say('vscode.commit-changes.it-is-one-commit-that-undo', { text: plan.text }) },
      commit,
    )

    return only === commit ? 'operation' : undefined
  }
  const only = say('vscode.commit-changes.only-the-files-the-last-operation')
  const everything = say('vscode.commit-changes.everything-git-shows-as-changed-more', { length: plan.extras.length })
  const choice = await vscode.window.showInformationMessage(
    say('vscode.commit-changes.monosmith-which-files-go-into-the'),
    {
      modal:  true,
      detail: say('vscode.commit-changes.the-first-choice-leaves-your-own', { text: plan.text, join: plan.extras.map((file) => `  ${file}`).join('\n') }),
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
      void vscode.window.showInformationMessage(say('vscode.commit-changes.monosmith', { part: outcome.lines.at(-1) ?? say('vscode.commit-changes.the-changes-are-committed') }))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.commit-changes.monosmith-nothing-was-committed', { message: outcome.message }))
      break
    }
    case 'declined': {
      break
    }
  }
}
