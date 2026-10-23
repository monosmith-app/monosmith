import { describe, expect, it } from 'vitest'
import type { WorkspaceSnapshot } from '../workspace/index.ts'
import { toTree } from './workspace-tree.mapper.ts'

const snapshot: WorkspaceSnapshot = {
  projects: [
    { name: 'web', dir: 'apps/web', kind: 'app', language: 'node', tasks: ['build'], deps: ['core'] },
    { name: 'core', dir: 'libs/core', kind: 'lib', language: 'go', tasks: ['test', 'lint'], deps: [] },
    { name: 'api', dir: 'apps/api', kind: 'app', language: 'go', tasks: [], deps: [] },
  ],
  affected: new Set(['core']),
}

describe('toTree', () => {
  it('groups by kind and sorts the projects by name', () => {
    const tree = toTree(snapshot, () => undefined)
    expect(tree.map((group) => group.label)).toEqual(['apps', 'libs'])
    expect(tree[0]?.children.map((project) => project.label)).toEqual(['api', 'web'])
  })

  it('marks the affected projects', () => {
    const core = toTree(snapshot, () => undefined)[1]?.children[0]
    expect(core).toMatchObject({ label: 'core', description: 'go · affected', affected: true })
  })

  it('puts the tasks under their project with the status of the last run', () => {
    const core = toTree(snapshot, (task) => (task === 'core:test' ? 'cached' : undefined))[1]?.children[0]
    expect(core?.children).toMatchObject([
      { id: 'task:core:test', project: 'core', task: 'test', status: 'cached', description: 'cached' },
      { id: 'task:core:lint', status: undefined, description: '' },
    ])
  })

  it('leaves out a kind with no projects and keeps an unusual one', () => {
    const tree = toTree({ projects: [{ name: 'x', dir: 'x', kind: 'tool', language: '', tasks: [], deps: [] }], affected: new Set() }, () => undefined)
    expect(tree.map((group) => group.label)).toEqual(['tool'])
  })
})
