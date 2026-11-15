import * as vscode from 'vscode'
import type { FixFindingOutcome } from './fix-finding.use-case.ts'
import { say } from '../words/index.ts'

/**
 * Asks one question of a fix. The first answer is marked as the suggested one.
 * @param question - what the person decides
 * @param labels - the answers, each saying what it does
 * @returns the index of the answer, or undefined when the person cancels
 */
export async function pickFixAnswer (question: string, labels: string[]): Promise<number | undefined> {
  const picked = await vscode.window.showQuickPick(
    labels.map((label, index) => ({ label: index === 0 ? `${label} (suggested)` : label, index })),
    { title: say('vscode.health.monosmith-fix'), placeHolder: question },
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
    say('vscode.health.monosmith', { description }),
    { modal: true, detail: say('vscode.health.nothing-is-changed-until-you-apply', { review }) },
    say('fixwizard.style.apply'),
  )

  return choice === say('fixwizard.style.apply')
}

/**
 * Says how fixing ended. It does not wait for the person to read it.
 * @param outcome - what `fixFinding` returned
 */
export function tellFixOutcome (outcome: FixFindingOutcome): void {
  switch (outcome.kind) {
    case 'fixed': {
      void vscode.window.showInformationMessage(say(outcome.incomplete ? 'vscode.health.monosmith-the-fix-was-applied-other' : 'vscode.health.monosmith-the-fix-was-applied'))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.health.monosmith-nothing-was-fixed', { message: outcome.message }))
      break
    }
    case 'nothing': {
      void vscode.window.showInformationMessage(say('vscode.health.monosmith-this-check-has-no-fix'))
      break
    }
    case 'declined': {
      break
    }
  }
}
