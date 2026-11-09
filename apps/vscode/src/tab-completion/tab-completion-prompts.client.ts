import * as vscode from 'vscode'
import { say } from '../words/index.ts'
import type { TabCompletionAnswer } from './offer-tab-completion.use-case.ts'

/** The one-time offer to set up tab completion. It writes nothing: the terminal asks before it does. */
export async function askTabCompletion (): Promise<TabCompletionAnswer | undefined> {
  const setup = say('vscode.tab-completion.set-up')
  const notNow = say('vscode.tab-completion.not-now')
  const never = say('vscode.tab-completion.dont-ask-again')
  const choice = await vscode.window.showInformationMessage(say('vscode.tab-completion.offer'), setup, notNow, never)
  if (choice === setup) return 'setup'
  if (choice === never) return 'never'

  return choice === notNow ? 'later' : undefined
}
