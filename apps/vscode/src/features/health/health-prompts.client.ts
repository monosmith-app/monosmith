import * as vscode from 'vscode'
import type { FixPreview } from './fix-preview.mapper.ts'
import { say } from '../words/index.ts'

/** Says smith is not running, so the doctor has nothing to check. */
export function tellNothingToCheck (): void {
  void vscode.window.showInformationMessage(say('vscode.health.monosmith-smith-is-not-running-so'))
}

/**
 * Says what the doctor found, with a button for the Output channel.
 * @param message - the one-line summary
 * @param problems - whether a check failed or warned
 * @param showOutput - brings the Output channel to the front
 */
export function tellDoctorSummary (message: string, problems: boolean, showOutput: () => void): void {
  const show = problems ? vscode.window.showWarningMessage : vscode.window.showInformationMessage
  void show(say('vscode.compose-stack.monosmith', { message }), say('vscode.health.show-output')).then((choice) => { if (choice === say('vscode.health.show-output')) showOutput() })
}

/** Says the doctor proposes no fix. */
export function tellNoFix (): void {
  void vscode.window.showInformationMessage(say('vscode.health.monosmith-doctor-proposes-no-fix-the'))
}

/**
 * Asks whether to apply the fixes the doctor proposes.
 * @param preview - the titles and the diffs of the fixes
 * @param folderName - the folder, when several run smith; undefined otherwise
 * @returns true on a yes
 */
export async function confirmApplyFixes (preview: FixPreview, folderName: string | undefined): Promise<boolean> {
  const choice = await vscode.window.showWarningMessage(
    say('vscode.health.monosmith-apply', { length: preview.titles.length, part: say(preview.titles.length === 1 ? 'vscode.health.fix' : 'vscode.health.fixes'), part2: folderName === undefined ? '' : say('vscode.health.in', { folderName }) }),
    { modal: true, detail: say('vscode.health.the-files-are-changed-as-shown', { text: preview.text }) },
    say('vscode.health.apply-fixes'),
  )

  return choice === say('vscode.health.apply-fixes')
}

/**
 * Says how the fixes ended.
 * @param outcome - `fixed` or `incomplete`; any other outcome says nothing
 * @param folderName - the folder, when several run smith; undefined otherwise
 */
export function tellFixesApplied (outcome: string, folderName: string | undefined): void {
  if (outcome === 'fixed') void vscode.window.showInformationMessage(say('vscode.health.monosmith-the-fixes-were-applied', { part: folderName === undefined ? '' : say('vscode.health.in', { folderName }) }))
  else if (outcome === 'incomplete') void vscode.window.showInformationMessage(say('vscode.health.monosmith-the-fixes-were-applied-some'))
}

/** Says a Health row must be chosen first. */
export function tellChooseCheck (): void {
  void vscode.window.showInformationMessage(say('vscode.health.monosmith-choose-a-check-in-the'))
}

/** Says the check of a Health row passes now. */
export function tellCheckPasses (): void {
  void vscode.window.showInformationMessage(say('vscode.health.monosmith-that-check-passes-now-refresh'))
}
