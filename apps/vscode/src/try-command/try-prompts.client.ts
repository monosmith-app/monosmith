import * as vscode from 'vscode'
import type { TryOutcome } from './try-command.use-case.ts'

/**
 * Asks which smith command to try.
 * @returns the words after `smith`, or undefined when the person cancels
 */
export async function askTryCommand (): Promise<string[] | undefined> {
  const text = await vscode.window.showInputBox({
    title:       'MonoSmith: try a command in a copy',
    prompt:      'Which smith command? It runs in a temporary copy first (add, remove, up, upgrade, derive, kind, eject, sync, bootstrap).',
    placeHolder: 'add package billing',
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
  const choice = await vscode.window.showInformationMessage(`MonoSmith: the try of "smith ${command.join(' ')}" passed. Run it for real?`, { modal: true }, 'Run it')

  return choice === 'Run it'
}

/**
 * Says how trying ended. It does not wait for the person to read it.
 * @param outcome - what `tryCommand` returned
 */
export function tellTryOutcome (outcome: TryOutcome): void {
  switch (outcome.kind) {
    case 'tried': {
      if (outcome.applied) void vscode.window.showInformationMessage('MonoSmith: the command ran for real.')
      else if (!outcome.passed) void vscode.window.showWarningMessage('MonoSmith: the try failed in the copy. Your files were not touched. The page says where.')
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: smith could not try that. ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
