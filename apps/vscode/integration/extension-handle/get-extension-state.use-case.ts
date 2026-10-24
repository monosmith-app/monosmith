import * as vscode from 'vscode'
import type { ExtensionStateApi } from '../../src/extension-api/index.ts'

/** The id the extension has: publisher and name of package.json. */
export const EXTENSION_ID = 'monosmith.monosmith-vscode'

/** Activates the extension and returns what it says about itself. */
export async function getExtensionState (): Promise<ExtensionStateApi> {
  const extension = vscode.extensions.getExtension<ExtensionStateApi>(EXTENSION_ID)
  if (extension === undefined) throw new Error(`the extension ${EXTENSION_ID} is not loaded`)

  return await extension.activate()
}
