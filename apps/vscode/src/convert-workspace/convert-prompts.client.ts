import * as vscode from 'vscode'
import type { ConvertOutcome } from './convert-workspace.contract.ts'

const TITLE = 'MonoSmith: convert this workspace'

/**
 * Asks the unticked "Commit this change so it can be undone" box. Nothing is changed.
 * @returns true when ticked, false when not, undefined when the person pressed Escape
 */
export async function askConvertCommit (): Promise<boolean | undefined> {
  const ticked = await vscode.window.showQuickPick(
    [{ label: 'Commit this change so it can be undone', description: 'unticked: smith commits nothing unless you tick this', picked: false }],
    { title: `${TITLE}: commit?`, canPickMany: true, placeHolder: 'Enter with nothing ticked leaves the change uncommitted' },
  )

  return ticked === undefined ? undefined : ticked.length > 0
}

/**
 * Shows the plan of `smith convert --dry-run`, which lists the files it deletes, and asks.
 * @param plan - the plan as smith printed it
 * @returns true when the person agrees
 */
export async function confirmConvert (plan: string): Promise<boolean> {
  const convert = 'Convert'
  const open = 'Open the plan'
  for (;;) {
    const choice = await vscode.window.showWarningMessage(
      `${TITLE}?`,
      { modal: true, detail: 'Nothing was changed yet. The plan lists the files smith deletes and the changes it makes. It is one click away.' },
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
      void vscode.window.showInformationMessage(`MonoSmith: the workspace is converted. ${outcome.committed ? 'smith committed it as one change that Undo reverts.' : 'Nothing was committed: review the files with git.'}`)
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was changed. ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
