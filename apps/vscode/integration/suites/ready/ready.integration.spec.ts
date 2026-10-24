import * as assert from 'node:assert/strict'
import * as vscode from 'vscode'
import { EXTENSION_ID, getExtensionState, waitUntil } from '../../extension-handle/index.ts'

/** A project row as the commands of the tree receive it. */
function projectRow (root: string, project: string): unknown {
  return { id: `project:${root}:${project}`, kind: 'project', label: project, root, projects: [project], project, children: [] }
}

describe('MonoSmith with a smith that runs', () => {
  it('starts smith in the folder and shows its projects', async () => {
    const state = await getExtensionState()
    await waitUntil('the view to be ready', () => state.viewState() === 'ready')
    assert.deepEqual(state.projects().sort((a, b) => a.localeCompare(b)), ['demo', 'lib'])
    assert.equal(state.roots().length, 1)
  })

  it('binds the keyboard shortcuts of running, re-running and cancelling', () => {
    const manifest = vscode.extensions.getExtension(EXTENSION_ID)?.packageJSON as { contributes: { keybindings: Array<{ command: string, key: string }> } }
    const keys = Object.fromEntries(manifest.contributes.keybindings.map((binding) => [binding.command, binding.key]))
    assert.equal(keys['monosmith.runTask'], 'ctrl+alt+m t')
    assert.equal(keys['monosmith.runAffected'], 'ctrl+alt+m a')
    assert.equal(keys['monosmith.rerunLast'], 'ctrl+alt+m r')
    assert.equal(keys['monosmith.cancelRun'], 'ctrl+alt+m c')
    assert.equal(keys['monosmith.runSelected'], 'enter')
  })

  it('runs the tasks of a project and records how they ended', async () => {
    const state = await getExtensionState()
    const [root = ''] = state.roots()
    await vscode.commands.executeCommand('monosmith.runAllTasks', projectRow(root, 'demo'))
    assert.equal(state.taskStatus('demo:build'), 'passed')
    assert.equal(state.taskStatus('demo:test'), 'passed')
    assert.equal(state.isRunning(), false)
  })

  it('repeats the last run with Re-run Last', async () => {
    const state = await getExtensionState()
    const before = state.taskStatus('demo:build')
    assert.equal(before, 'passed')
    await vscode.commands.executeCommand('monosmith.rerunLast')
    assert.equal(state.taskStatus('demo:build'), 'passed')
  })

  it('cancels a run that does not end by itself', async () => {
    const state = await getExtensionState()
    const [root = ''] = state.roots()
    const finished = vscode.commands.executeCommand('monosmith.runAllTasks', projectRow(root, 'lib'))
    await waitUntil('the run to start', () => state.isRunning())
    await vscode.commands.executeCommand('monosmith.cancelRun')
    await finished
    await waitUntil('the run to end', () => !state.isRunning())
    assert.equal(state.taskStatus('lib:hang'), 'failed')
  })

  it('refreshes without error', async () => {
    const state = await getExtensionState()
    await vscode.commands.executeCommand('monosmith.refresh')
    assert.equal(state.viewState(), 'ready')
  })
})
