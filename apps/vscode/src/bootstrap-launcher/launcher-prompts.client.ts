import * as vscode from 'vscode'
import type { LauncherOutcome } from './write-launcher.use-case.ts'
import { say } from '../words/index.ts'

/**
 * Asks whether to write the launcher files.
 * @param folder - the name of the workspace folder
 * @param plan - the plan smith printed: the files it writes, with a diff of one that exists
 * @returns whether the person agreed
 */
export async function confirmLauncher (folder: string, plan: string): Promise<boolean> {
  const choice = await vscode.window.showInformationMessage(
    say('vscode.bootstrap-launcher.monosmith-write-the-smith-launcher-in', { folder }),
    { modal: true, detail: say('vscode.bootstrap-launcher.they-run-the-smith-version-that', { plan }) },
    say('vscode.agent-install.write'),
  )

  return choice === say('vscode.agent-install.write')
}

/**
 * Says how writing ended. It does not wait for the person to read it.
 * @param outcome - what `writeLauncher` returned
 */
export function tellLauncherOutcome (outcome: LauncherOutcome): void {
  switch (outcome.kind) {
    case 'written': {
      void vscode.window.showInformationMessage(say('vscode.bootstrap-launcher.monosmith-the-launcher-is-written-in'))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.bootstrap-launcher.monosmith-the-launcher-was-not-written', { message: outcome.message }))
      break
    }
    case 'declined': {
      break
    }
  }
}
