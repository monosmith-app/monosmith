import * as vscode from 'vscode'
import type { FolderState } from '../folder-session/index.ts'

/**
 * Asks which folder to work in, or takes the only one.
 * @param candidates - the folders that run smith
 * @returns the folder, or undefined when there is none or the person pressed Escape
 */
export async function pickFolder (candidates: readonly FolderState[]): Promise<FolderState | undefined> {
  if (candidates.length <= 1) return candidates[0]
  const picked = await vscode.window.showQuickPick(candidates.map((state) => ({ label: state.name, description: state.root, state })), { title: 'Choose a folder', placeHolder: 'Folder' })

  return picked?.state
}

/** Says smith does not run in any folder yet. */
export function tellNoFolder (): void {
  void vscode.window.showInformationMessage('MonoSmith: smith is not running in a folder, so there is nothing to do yet. Open a folder with a smith.toml and wait for the MonoSmith view to say it is ready.')
}

/**
 * Says a command that writes files needs a trusted workspace.
 * @param what - the words that name the command, such as "committing"
 */
export function tellNeedsTrust (what: string): void {
  void vscode.window.showWarningMessage(`MonoSmith: ${what} writes files and needs a trusted workspace. Trust the workspace, then try again.`)
}

/**
 * Asks which of several folders, or takes the only one.
 * @param candidates - the folders to choose from
 * @returns the folder, or undefined when there is none or the person pressed Escape
 */
export async function chooseFolder<T extends { root: string, name: string }> (candidates: readonly T[]): Promise<T | undefined> {
  if (candidates.length <= 1) return candidates[0]
  const picked = await vscode.window.showQuickPick(candidates.map((candidate) => ({ label: candidate.name, description: candidate.root, candidate })), { placeHolder: 'Which folder?' })

  return picked?.candidate
}
