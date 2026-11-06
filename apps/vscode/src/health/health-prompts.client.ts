import * as vscode from 'vscode'
import type { FixPreview } from './fix-preview.mapper.ts'

/** Says smith is not running, so the doctor has nothing to check. */
export function tellNothingToCheck (): void {
  void vscode.window.showInformationMessage('MonoSmith: smith is not running, so there is nothing to check.')
}

/**
 * Says what the doctor found, with a button for the Output channel.
 * @param message - the one-line summary
 * @param problems - whether a check failed or warned
 * @param showOutput - brings the Output channel to the front
 */
export function tellDoctorSummary (message: string, problems: boolean, showOutput: () => void): void {
  const show = problems ? vscode.window.showWarningMessage : vscode.window.showInformationMessage
  void show(`MonoSmith: ${message}`, 'Show output').then((choice) => { if (choice === 'Show output') showOutput() })
}

/** Says the doctor proposes no fix. */
export function tellNoFix (): void {
  void vscode.window.showInformationMessage('MonoSmith: doctor proposes no fix. The remedies listed under a check say what to do by hand.')
}

/**
 * Asks whether to apply the fixes the doctor proposes.
 * @param preview - the titles and the diffs of the fixes
 * @param folderName - the folder, when several run smith; undefined otherwise
 * @returns true on a yes
 */
export async function confirmApplyFixes (preview: FixPreview, folderName: string | undefined): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    `MonoSmith: apply ${preview.titles.length} ${preview.titles.length === 1 ? 'fix' : 'fixes'}${folderName === undefined ? '' : ` in ${folderName}`}?`,
    { modal: true, detail: `${preview.text}\n\nThe files are changed as shown. Review them with git afterwards.` },
    'Apply fixes',
  )

  return choice === 'Apply fixes'
}

/**
 * Says how the fixes ended.
 * @param outcome - `fixed` or `incomplete`; any other outcome says nothing
 * @param folderName - the folder, when several run smith; undefined otherwise
 */
export function tellFixesApplied (outcome: string, folderName: string | undefined): void {
  if (outcome === 'fixed') void vscode.window.showInformationMessage(`MonoSmith: the fixes were applied${folderName === undefined ? '' : ` in ${folderName}`}.`)
  else if (outcome === 'incomplete') void vscode.window.showInformationMessage('MonoSmith: the fixes were applied. Some checks still fail and need their remedy by hand; the Health view lists them.')
}

/** Says a Health row must be chosen first. */
export function tellChooseCheck (): void {
  void vscode.window.showInformationMessage('MonoSmith: choose a check in the Health view first.')
}

/** Says the check of a Health row passes now. */
export function tellCheckPasses (): void {
  void vscode.window.showInformationMessage('MonoSmith: that check passes now. Refresh the Health view.')
}
