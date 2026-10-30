import * as assert from 'node:assert/strict'
import { existsSync } from 'node:fs'
import { join } from 'node:path'
import * as vscode from 'vscode'
import { getExtensionState, waitUntil } from '../../extension-handle/index.ts'

describe('MonoSmith in a folder with files and no recognised tool', () => {
  it('offers to bring the folder in, without running smith', async () => {
    const state = await getExtensionState()
    await waitUntil('the view to offer the folder', () => state.viewState() === 'unrecognised')
    assert.deepEqual(state.roots(), [])
    assert.deepEqual(state.offers(), [])
  })

  it('writes nothing by itself', () => {
    const root = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath ?? ''
    assert.equal(existsSync(join(root, 'smith.toml')), false)
  })
})
