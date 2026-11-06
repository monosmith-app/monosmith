import * as vscode from 'vscode'
import type { ConvertOutcome } from './convert-workspace.contract.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.convert-workspace.monosmith-convert-this-workspace')

/**
 * Asks the unticked "Commit this change so it can be undone" box. Nothing is changed.
 * @returns true when ticked, false when not, undefined when the person pressed Escape
 */
export async function askConvertCommit (): Promise<boolean | undefined> {
  const ticked = await vscode.window.showQuickPick(
    [{ label: say('dialog.commit.label'), description: say('vscode.adopt-offer.unticked-smith-commits-nothing-unless-you'), picked: false }],
    { title: `${TITLE}: commit?`, canPickMany: true, placeHolder: say('vscode.adopt-offer.enter-with-nothing-ticked-leaves-the') },
  )

  return ticked === undefined ? undefined : ticked.length > 0
}

/**
 * Shows the plan of `smith convert --dry-run`, which lists the files it deletes, and asks.
 * @param plan - the plan as smith printed it
 * @returns true when the person agrees
 */
export async function confirmConvert (plan: string): Promise<boolean> {
  const convert = say('vscode.convert-workspace.convert')
  const open = say('vscode.add-project.open-the-plan')
  for (;;) {
    const choice = await vscode.window.showWarningMessage(
      `${TITLE}?`,
      { modal: true, detail: say('vscode.convert-workspace.nothing-was-changed-yet-the-plan') },
      convert,
      open,
    )
    if (choice !== open) return choice === convert
    const document = await vscode.workspace.openTextDocument({ language: 'plaintext', content: `${plan}\n` })
    await vscode.window.showTextDocument(document, { preview: true })
  }
}

/**
 * Says how converting ended. It does not wait for the person to read it.
 * @param outcome - what `convertWorkspace` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellConvertOutcome (outcome: ConvertOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'converted': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(say('vscode.convert-workspace.monosmith-the-workspace-is-converted', { part: say(outcome.committed ? 'vscode.convert-workspace.smith-committed-it-as-one-change' : 'vscode.convert-workspace.nothing-was-committed-review-the-files') }))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.convert-workspace.monosmith-nothing-was-changed', { message: outcome.message }))
      break
    }
    case 'cancelled': {
      break
    }
  }
}

/** Says no Nx or Turborepo workspace is open. */
export function tellNoConvertible (): void {
  void vscode.window.showInformationMessage(say('vscode.convert-workspace.monosmith-no-nx-or-turborepo-workspace'))
}
