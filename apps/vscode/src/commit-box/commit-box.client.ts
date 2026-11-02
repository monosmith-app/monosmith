import * as vscode from 'vscode'

/**
 * Asks the unticked "Commit this change so it can be undone" box before the confirmation. smith
 * commits only when it is ticked, and then only the files of the operation. Nothing is changed.
 * @param title - what the change is, for the title of the box
 * @returns true when ticked, false when not, undefined when the person pressed Escape
 */
export async function askCommitBox (title: string): Promise<boolean | undefined> {
  const ticked = await vscode.window.showQuickPick(
    [{ label: 'Commit this change so it can be undone', description: 'unticked: smith commits nothing unless you tick this', picked: false }],
    { title: `MonoSmith: ${title.replace(/…$/, '')}: commit?`, canPickMany: true, placeHolder: 'Enter with nothing ticked leaves the change uncommitted' },
  )

  return ticked === undefined ? undefined : ticked.length > 0
}
