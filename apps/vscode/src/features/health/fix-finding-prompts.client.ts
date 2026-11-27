import * as vscode from 'vscode'
import type { FixFindingOutcome } from './fix-finding.use-case.ts'
import { say } from '../../shared/words/index.ts'

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
 * Step 2 of the fix: asks whether to apply what the review shows. The command line it stands for is part of the question.
 * @param review - smith's review: diffs, commands and the plan each command prints
 * @param description - the fix
 * @param shownBeside - true when the review is open beside the code, so the question only points at it
 * @returns whether the person agreed
 */
export async function confirmFix (review: string, description: string, shownBeside = false): Promise<boolean> {
  const line = say('tui.detail.or_in_a_terminal_smith_doctor', { check: description })
  const choice = await vscode.window.showWarningMessage(
    say('vscode.health.monosmith', { description }),
    { modal: true, detail: shownBeside ? say('vscode.health.review-is-open-beside', { line }) : say('vscode.health.nothing-is-changed-until-you-apply', { review }) },
    say('fixwizard.style.apply'),
  )

  return choice === say('fixwizard.style.apply')
}

/**
 * Step 3 of the fix: says whether the check passes now that the fix is applied, that nothing was
 * committed, and offers to check again.
 * @param outcome - the applied fix
 * @param still - true when doctor still reports the check
 */
export function tellFixChecked (outcome: Extract<FixFindingOutcome, { kind: 'fixed' }>, still: boolean): void {
  const state = say(still ? 'healthfix.still' : 'healthfix.passes')
  const message = say('vscode.health.fix-checked', { state, committed: say('tui.capture.not_committed_review_with_git_status') })
  const again = say('healthfix.step.check')
  const shown = still || outcome.incomplete ? vscode.window.showWarningMessage(message, again) : vscode.window.showInformationMessage(message, again)
  void shown.then(async (choice) => { if (choice === again) await vscode.commands.executeCommand('monosmith.health.refresh') })
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
