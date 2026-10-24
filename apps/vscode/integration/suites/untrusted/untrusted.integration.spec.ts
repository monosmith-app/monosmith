import * as assert from 'node:assert/strict'
import * as vscode from 'vscode'
import { getExtensionState, waitUntil } from '../../extension-handle/index.ts'

describe('MonoSmith in a workspace that is not trusted', () => {
  it('is running in Restricted Mode', () => {
    assert.equal(vscode.workspace.isTrusted, false)
  })

  it('shows the projects with the smith of the PATH and ignores the workspace setting', async () => {
    const state = await getExtensionState()
    await waitUntil('the view to be ready', () => state.viewState() === 'ready')
    assert.deepEqual(state.projects().sort((a, b) => a.localeCompare(b)), ['demo', 'lib'])
  })

  it('does not run tasks', async () => {
    const state = await getExtensionState()
    const [root] = state.roots()
    // The command waits for the person to answer the warning, so it is not awaited.
    void vscode.commands.executeCommand('monosmith.runAllTasks', { id: 'p', kind: 'project', label: 'demo', root, projects: ['demo'], project: 'demo', children: [] })
    await new Promise((resolve) => setTimeout(resolve, 1500))
    assert.equal(state.isRunning(), false)
    assert.equal(state.taskStatus('demo:build'), undefined)
  })
})
