import * as vscode from 'vscode'
import type { TryOutcome } from './try-command.use-case.ts'
import { say } from '../words/index.ts'

/**
 * Asks which smith command to try.
 * @returns the words after `smith`, or undefined when the person cancels
 */
export async function askTryCommand (): Promise<string[] | undefined> {
  const text = await vscode.window.showInputBox({
    title:       say('vscode.try-command.monosmith-try-a-command-in-a'),
    prompt:      say('vscode.try-command.which-smith-command-it-runs-in'),
    placeHolder: say('guide.actions.add_package_billing'),
  })
  const words = text?.trim().split(/\s+/).filter(Boolean)

  return words === undefined || words.length === 0 ? undefined : words
}

/**
 * Asks whether to run a command for real after a try that passed.
 * @param command - the words after `smith`
 * @returns whether the person agreed
 */
export async function offerTryApply (command: string[]): Promise<boolean> {
  const choice = await vscode.window.showInformationMessage(say('vscode.try-command.monosmith-the-try-of-smith-passed', { join: command.join(' ') }), { modal: true }, say('vscode.compose-stack.run-it'))

  return choice === say('vscode.compose-stack.run-it')
}

/**
 * Says how trying ended. It does not wait for the person to read it.
 * @param outcome - what `tryCommand` returned
 */
export function tellTryOutcome (outcome: TryOutcome): void {
  switch (outcome.kind) {
    case 'tried': {
      if (outcome.applied) void vscode.window.showInformationMessage(say('vscode.try-command.monosmith-the-command-ran-for-real'))
      else if (!outcome.passed) void vscode.window.showWarningMessage(say('vscode.try-command.monosmith-the-try-failed-in-the'))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.try-command.monosmith-smith-could-not-try-that', { message: outcome.message }))
      break
    }
    case 'cancelled': {
      break
    }
  }
}
