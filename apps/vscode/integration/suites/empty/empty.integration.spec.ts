import * as assert from 'node:assert/strict'
import { readdirSync } from 'node:fs'
import * as vscode from 'vscode'
import { getExtensionState, waitUntil } from '../../extension-handle/index.ts'

describe('MonoSmith in an empty folder', () => {
  it('offers a new workspace, without running smith', async () => {
    const state = await getExtensionState()
    await waitUntil('the view to offer a new workspace', () => state.viewState() === 'empty')
    assert.deepEqual(state.roots(), [])
    assert.deepEqual(state.offers(), [])
  })

  it('has the command that creates it, and writes nothing by itself', async () => {
    const commands = await vscode.commands.getCommands(true)
    assert.ok(commands.includes('monosmith.new'))
    const root = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath ?? ''
    assert.deepEqual(readdirSync(root).filter((name) => name !== '.vscode'), [])
  })
})
