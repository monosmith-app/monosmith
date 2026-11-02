import * as vscode from 'vscode'
import type { FixFindingOutcome } from './fix-finding.use-case.ts'

/**
 * Asks one question of a fix. The first answer is marked as the suggested one.
 * @param question - what the person decides
 * @param labels - the answers, each saying what it does
 * @returns the index of the answer, or undefined when the person cancels
 */
export async function pickFixAnswer (question: string, labels: string[]): Promise<number | undefined> {
  const picked = await vscode.window.showQuickPick(
    labels.map((label, index) => ({ label: index === 0 ? `${label} (suggested)` : label, index })),
    { title: 'MonoSmith: fix', placeHolder: question },
  )

  return picked?.index
}

/**
 * Shows what the fix would change and asks.
 * @param review - smith's review: diffs, commands and the plan each command prints
 * @param description - the fix
 * @returns whether the person agreed
 */
export async function confirmFix (review: string, description: string): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    `MonoSmith: ${description}?`,
    { modal: true, detail: `${review}\n\nNothing is changed until you apply it. smith never commits unasked.` },
    'Apply',
  )

  return choice === 'Apply'
}

/**
 * Says how fixing ended. It does not wait for the person to read it.
 * @param outcome - what `fixFinding` returned
 */
export function tellFixOutcome (outcome: FixFindingOutcome): void {
  switch (outcome.kind) {
    case 'fixed': {
      void vscode.window.showInformationMessage(outcome.incomplete ? 'MonoSmith: the fix was applied. Other checks still fail; the Health view lists them.' : 'MonoSmith: the fix was applied.')
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was fixed. ${outcome.message}`)
      break
    }
    case 'nothing': {
      void vscode.window.showInformationMessage('MonoSmith: this check has no fix. Its remedy says what to do.')
      break
    }
    case 'declined': {
      break
    }
  }
}
