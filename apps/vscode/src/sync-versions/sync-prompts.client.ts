import * as vscode from 'vscode'
import type { SyncOutcome, SyncRule } from './sync-versions.use-case.ts'
import { say } from '../words/index.ts'

/**
 * Asks which version wins when projects declare a dependency at different
 * versions. Each answer says what it does; nothing is written yet.
 * @returns the rule, or undefined when the person pressed Escape
 */
export async function chooseSyncRule (): Promise<SyncRule | undefined> {
  const picked = await vscode.window.showQuickPick<vscode.QuickPickItem & { rule: SyncRule }>([
    { label: say('resultdialog.version.installed.label'), detail: say('vscode.sync-versions.the-version-the-package-manager-has'), rule: 'installed' },
    { label: say('resultdialog.version.highest.label'), detail: say('resultdialog.version.highest.hint'), rule: 'highest' },
    { label: say('resultdialog.version.lowest.label'), detail: say('resultdialog.version.lowest.hint'), rule: 'lowest' },
  ], { title: say('vscode.sync-versions.monosmith-which-version-wins'), placeHolder: say('resultdialog.foot.version') })

  return picked?.rule
}

/**
 * Shows the sync plan and asks.
 * @param plan - the text of `smith sync --dry-run`
 * @param rule - the rule the plan was made with
 * @returns whether the person agreed
 */
export async function confirmSync (plan: string, rule: SyncRule): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    say('vscode.sync-versions.monosmith-write-these-versions-wins', { rule }),
    { modal: true, detail: say('vscode.sync-versions.nothing-is-committed', { plan }) },
    say('vscode.sync-versions.write-them'),
  )

  return choice === say('vscode.sync-versions.write-them')
}

/**
 * Says how syncing ended. It does not wait for the person to read it.
 * @param outcome - what `syncVersions` returned
 */
export function tellSyncOutcome (outcome: SyncOutcome): void {
  switch (outcome.kind) {
    case 'synced': {
      void vscode.window.showInformationMessage(say('vscode.sync-versions.monosmith-the-versions-are-written-nothing'))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.sync-versions.monosmith-nothing-was-synced', { message: outcome.message }))
      break
    }
    case 'declined': {
      break
    }
  }
}
