import * as vscode from 'vscode'
import type { SyncOutcome, SyncRule } from './sync-versions.use-case.ts'

/**
 * Asks which version wins when projects declare a dependency at different
 * versions. Each answer says what it does; nothing is written yet.
 * @returns the rule, or undefined when the person pressed Escape
 */
export async function chooseSyncRule (): Promise<SyncRule | undefined> {
  const picked = await vscode.window.showQuickPick<vscode.QuickPickItem & { rule: SyncRule }>([
    { label: 'The installed version (smith\'s pick)', detail: 'the version the package manager has installed', rule: 'installed' },
    { label: 'The highest declared', detail: 'projects on an older version move up', rule: 'highest' },
    { label: 'The lowest declared', detail: 'projects on a newer version move down', rule: 'lowest' },
  ], { title: 'MonoSmith: which version wins?', placeHolder: 'You see the plan before anything is written. One dependency at a time: smith sync --to <name>=<version>.' })

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
    `MonoSmith: write these versions (${rule} wins)?`,
    { modal: true, detail: `${plan}\n\nNothing is committed.` },
    'Write them',
  )

  return choice === 'Write them'
}

/**
 * Says how syncing ended. It does not wait for the person to read it.
 * @param outcome - what `syncVersions` returned
 */
export function tellSyncOutcome (outcome: SyncOutcome): void {
  switch (outcome.kind) {
    case 'synced': {
      void vscode.window.showInformationMessage('MonoSmith: the versions are written. Nothing is committed: review the change with git.')
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was synced. ${outcome.message}`)
      break
    }
    case 'declined': {
      break
    }
  }
}
