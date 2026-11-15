import * as vscode from 'vscode'
import type { FolderState } from '../folder-session/index.ts'
import { say } from '../words/index.ts'

/**
 * Asks which folder to work in, or takes the only one.
 * @param candidates - the folders that run smith
 * @returns the folder, or undefined when there is none or the person pressed Escape
 */
export async function pickFolder (candidates: readonly FolderState[]): Promise<FolderState | undefined> {
  if (candidates.length <= 1) return candidates[0]
  const picked = await vscode.window.showQuickPick(candidates.map((state) => ({ label: state.name, description: state.root, state })), { title: say('vscode.command-host.choose-a-folder'), placeHolder: say('vscode.command-host.folder') })

  return picked?.state
}

/** Says smith does not run in any folder yet. */
export function tellNoFolder (): void {
  void vscode.window.showInformationMessage(say('vscode.command-host.monosmith-smith-is-not-running-in'))
}

/**
 * Says a command that writes files needs a trusted workspace.
 * @param what - the words that name the command, such as "committing"
 */
export function tellNeedsTrust (what: string): void {
  void vscode.window.showWarningMessage(say('vscode.command-host.monosmith-writes-files-and-needs-a', { what }))
}

/**
 * Asks which of several folders, or takes the only one.
 * @param candidates - the folders to choose from
 * @returns the folder, or undefined when there is none or the person pressed Escape
 */
export async function chooseFolder<T extends { root: string, name: string }> (candidates: readonly T[]): Promise<T | undefined> {
  if (candidates.length <= 1) return candidates[0]
  const picked = await vscode.window.showQuickPick(candidates.map((candidate) => ({ label: candidate.name, description: candidate.root, candidate })), { placeHolder: say('vscode.command-host.which-folder') })

  return picked?.candidate
}
