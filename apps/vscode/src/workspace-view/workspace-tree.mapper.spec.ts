import { describe, expect, it } from 'vitest'
import type { WorkspaceSnapshot } from '../workspace/index.ts'
import { toTree, type TreeFolder } from './workspace-tree.mapper.ts'

const snapshot: WorkspaceSnapshot = {
  projects: [
    { name: 'web', dir: 'apps/web', kind: 'app', language: 'node', tasks: ['build'], deps: ['core'] },
    { name: 'core', dir: 'libs/core', kind: 'lib', language: 'go', tasks: ['test', 'lint'], deps: [] },
    { name: 'api', dir: 'apps/api', kind: 'app', language: 'go', tasks: [], deps: [] },
  ],
  affected: new Set(['core']),
}
const single: TreeFolder[] = [{ root: '/repo', name: 'repo', snapshot }]
const none = (): undefined => undefined

describe('toTree', () => {
  it('groups by kind and sorts the projects by name', () => {
    const tree = toTree(single, none)
    expect(tree.map((group) => group.label)).toEqual(['apps', 'libs'])
    expect(tree[0]?.children.map((project) => project.label)).toEqual(['api', 'web'])
  })

  it('marks the affected projects', () => {
    const core = toTree(single, none)[1]?.children[0]
    expect(core).toMatchObject({ label: 'core', description: 'go · affected', affected: true })
  })

  it('puts the tasks under their project with the status of the last run', () => {
    const core = toTree(single, (_root, task) => (task === 'core:test' ? 'cached' : undefined))[1]?.children[0]
    expect(core?.children).toMatchObject([
      { id: 'task:/repo:core:test', project: 'core', task: 'test', status: 'cached', description: 'cached', root: '/repo' },
      { id: 'task:/repo:core:lint', status: undefined, description: '' },
    ])
  })

  it('leaves out a kind with no projects and keeps an unusual one', () => {
    const odd: TreeFolder[] = [{ root: '/r', name: 'r', snapshot: { projects: [{ name: 'x', dir: 'x', kind: 'tool', language: '', tasks: [], deps: [] }], affected: new Set() } }]
    expect(toTree(odd, none).map((group) => group.label)).toEqual(['tool'])
  })

  it('says what a project and a task are to a screen reader', () => {
    const core = toTree(single, (_root, task) => (task === 'core:test' ? 'failed' : undefined))[1]?.children[0]
    expect(core?.spoken).toBe('core, go, affected by the changes, 2 tasks')
    expect(core?.children[0]?.spoken).toBe('core test, failed. Press Enter to run.')
    expect(core?.children[1]?.spoken).toContain('not run yet')
  })

  it('lists the projects a group and a project stand for, to run all their tasks', () => {
    const tree = toTree(single, none)
    expect(tree[0]?.projects).toEqual(['api', 'web'])
    expect(tree[0]?.children[1]?.projects).toEqual(['web'])
  })

  it('is empty when no folder has a running smith, so the welcome content shows', () => {
    expect(toTree([{ root: '/r', name: 'r', note: 'smith not found' }], none)).toEqual([])
    expect(toTree([], none)).toEqual([])
  })

  describe('with several folders', () => {
    const many: TreeFolder[] = [
      { root: '/a', name: 'a', snapshot },
      { root: '/b', name: 'b', note: 'smith not found' },
    ]

    it('has a row per folder, with groups under the ones that run', () => {
      const tree = toTree(many, none)
      expect(tree.map((node) => [node.kind, node.label])).toEqual([['folder', 'a'], ['folder', 'b']])
      expect(tree[0]?.children.map((group) => group.label)).toEqual(['apps', 'libs'])
    })

    it('says why a folder has nothing', () => {
      expect(toTree(many, none)[1]).toMatchObject({ description: 'smith not found', children: [] })
    })

    it('keeps the ids of two folders apart', () => {
      const both: TreeFolder[] = [{ root: '/a', name: 'a', snapshot }, { root: '/b', name: 'b', snapshot }]
      const ids = toTree(both, none).flatMap((folder) => folder.children.flatMap((group) => group.children.map((project) => project.id)))
      expect(new Set(ids).size).toBe(ids.length)
    })

    it('hands each task the status of its own folder', () => {
      const both: TreeFolder[] = [{ root: '/a', name: 'a', snapshot }, { root: '/b', name: 'b', snapshot }]
      const tree = toTree(both, (root, task) => (root === '/a' && task === 'web:build' ? 'passed' : undefined))
      const statusIn = (index: number) => tree[index]?.children[0]?.children[1]?.children[0]?.status
      expect(statusIn(0)).toBe('passed')
      expect(statusIn(1)).toBeUndefined()
    })
  })
})
