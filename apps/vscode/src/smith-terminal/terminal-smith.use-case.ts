import { existsSync } from 'node:fs'
import * as vscode from 'vscode'
import { locateSmith } from '../smith-client/index.ts'
import { isWorkspaceTrusted } from '../workspace-trust/index.ts'
import { tellSmithNotInstalled } from './terminal-prompts.client.ts'

/**
 * Finds the smith a terminal in the folder would run.
 * @param root - the folder
 * @returns the path of smith, or undefined after saying it is not installed
 */
export async function smithForTerminal (root: string): Promise<string | undefined> {
  const found = locateSmith({
    setting:       vscode.workspace.getConfiguration('monosmith').get<string>('path', ''),
    workspaceRoot: root,
    trusted:       isWorkspaceTrusted(),
    platform:      process.platform,
    pathVariable:  process.env.PATH ?? '',
    exists:        existsSync,
  })
  if (found.location === undefined) {
    await tellSmithNotInstalled()

    return undefined
  }

  return found.location.executable
}
