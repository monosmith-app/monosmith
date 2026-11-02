import * as vscode from 'vscode'
import type { LauncherOutcome } from './write-launcher.use-case.ts'

/**
 * Asks whether to write the launcher files.
 * @param folder - the name of the workspace folder
 * @param plan - the plan smith printed: the files it writes, with a diff of one that exists
 * @returns whether the person agreed
 */
export async function confirmLauncher (folder: string, plan: string): Promise<boolean> {
  const choice = await vscode.window.showInformationMessage(
    `MonoSmith: write the smith launcher in ${folder}?`,
    { modal: true, detail: `${plan}\n\nThey run the smith version that smith.toml pins, so a laptop and CI use the same one. A launcher that exists is replaced. Commit them.` },
    'Write',
  )

  return choice === 'Write'
}

/**
 * Says how writing ended. It does not wait for the person to read it.
 * @param outcome - what `writeLauncher` returned
 */
export function tellLauncherOutcome (outcome: LauncherOutcome): void {
  switch (outcome.kind) {
    case 'written': {
      void vscode.window.showInformationMessage('MonoSmith: the launcher is written in tools/smith. Commit it.')
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: the launcher was not written. ${outcome.message}`)
      break
    }
    case 'declined': {
      break
    }
  }
}
