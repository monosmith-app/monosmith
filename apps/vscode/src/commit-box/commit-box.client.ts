import * as vscode from 'vscode'
import { say } from '../words/index.ts'

/**
 * Asks the unticked "Commit this change so it can be undone" box before the confirmation. smith
 * commits only when it is ticked, and then only the files of the operation. Nothing is changed.
 * @param title - what the change is, for the title of the box
 * @returns true when ticked, false when not, undefined when the person pressed Escape
 */
export async function askCommitBox (title: string): Promise<boolean | undefined> {
  const ticked = await vscode.window.showQuickPick(
    [{ label: say('dialog.commit.label'), description: say('vscode.adopt-offer.unticked-smith-commits-nothing-unless-you'), picked: false }],
    { title: say('vscode.commit-box.monosmith-commit', { replace: title.replace(/…$/, '') }), canPickMany: true, placeHolder: say('vscode.adopt-offer.enter-with-nothing-ticked-leaves-the') },
  )

  return ticked === undefined ? undefined : ticked.length > 0
}
