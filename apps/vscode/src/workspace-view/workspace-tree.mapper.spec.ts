import { describe, expect, it } from 'vitest'
import type { WorkspaceSnapshot } from '../workspace/index.ts'
import { countProjects, describeAffected, describeFilter, formatDuration, toTree, type TaskFacts, type TreeFolder } from './workspace-tree.mapper.ts'

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
const only = (task: string, facts: TaskFacts) => (_root: string, id: string): TaskFacts | undefined => (id === task ? facts : undefined)
const coreTest = (facts: TaskFacts) => only('core:test', facts)

describe('toTree', () => {
  it('groups by kind and sorts the projects by name', () => {
    const tree = toTree(single, none)
    expect(tree.map((group) => group.label)).toEqual(['apps', 'libs'])
    expect(tree[0]?.children.map((project) => project.label)).toEqual(['api', 'web'])
  })

  it('gives each kind its icon', () => {
    const tree = toTree(single, none)
    expect(tree.map((group) => group.icon)).toEqual(['window', 'library'])
    expect(tree[1]?.children[0]?.icon).toBe('library')
  })

  it('says the language beside a project and marks an affected one with a dot, without the word', () => {
    const core = toTree(single, none)[1]?.children[0]
    expect(core).toMatchObject({ label: 'core', description: 'go', affected: true, dot: true })
    expect(core?.tooltip).toContain('Affected by the changes.')
    expect(toTree(single, none)[0]?.children[1]).toMatchObject({ label: 'web', description: 'node', affected: false, dot: false })
  })

  it('shows no dot at all when every project is affected, since it would tell nothing', () => {
    const every: TreeFolder[] = [{ root: '/r', name: 'r', snapshot: { ...snapshot, affected: new Set(['web', 'core', 'api']) } }]
    const projects = toTree(every, none).flatMap((group) => group.children)
    expect(projects.every((project) => project.affected === true && project.dot === false)).toBe(true)
  })

  it('puts the tasks under their project, with only what the icon does not say', () => {
    const facts = new Map<string, TaskFacts>([
      ['core:test', { status: 'passed', durationMs: 1234, hasLog: true }],
      ['core:lint', { status: 'failed', reason: 'exited with code 1' }],
    ])
    const core = toTree(single, (_root, task) => facts.get(task))[1]?.children[0]
    expect(core?.children).toMatchObject([
      { id: 'task:/repo:core:test', task: 'test', status: 'passed', description: '1.2s', icon: 'pass', hasLog: true, root: '/repo' },
      { id: 'task:/repo:core:lint', status: 'failed', description: 'exited with code 1', icon: 'error' },
    ])
  })

  it('does not repeat the status word of a cached task, and a task never run has the play icon', () => {
    const core = toTree(single, coreTest({ status: 'cached' }))[1]?.children[0]
    expect(core?.children[0]).toMatchObject({ status: 'cached', description: '', icon: 'history' })
    expect(core?.children[1]).toMatchObject({ description: '', icon: 'play' })
    expect(core?.children[1]?.status).toBeUndefined()
  })

  it('tells the last result, the duration and the cache hit in the tooltip', () => {
    const children = toTree(single, (_root, task) => (task === 'core:test' ? { status: 'cached' } : { status: 'passed', durationMs: 2500, time: '2026-10-20T06:05:59Z' }))[1]?.children[0]?.children
    expect(children?.[0]?.tooltip).toContain('Cache hit')
    expect(children?.[1]?.tooltip).toContain('Passed in 2.5s.')
    expect(children?.[1]?.tooltip).toContain('Last run:')
  })

  it('leaves out a kind with no projects and keeps an unusual one', () => {
    const odd: TreeFolder[] = [{ root: '/r', name: 'r', snapshot: { projects: [{ name: 'x', dir: 'x', kind: 'tool', language: '', tasks: [], deps: [] }], affected: new Set() } }]
    expect(toTree(odd, none).map((group) => group.label)).toEqual(['tool'])
  })

  it('says what a project and a task are to a screen reader', () => {
    const core = toTree(single, coreTest({ status: 'failed', durationMs: 900 }))[1]?.children[0]
    expect(core?.spoken).toBe('core, go, affected by the changes, 2 tasks')
    expect(core?.children[0]?.spoken).toBe('core test, failed in 900ms. Press Enter to run.')
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

  describe('with a filter', () => {
    it('keeps the projects the text names, with all their tasks', () => {
      const tree = toTree(single, none, { text: 'COR', affectedOnly: false })
      expect(tree.map((group) => group.label)).toEqual(['libs'])
      expect(tree[0]?.children[0]).toMatchObject({ label: 'core', reveal: false })
      expect(tree[0]?.children[0]?.children).toHaveLength(2)
    })

    it('keeps a project for a task the text names, with that task only, open', () => {
      const tree = toTree(single, none, { text: 'lint', affectedOnly: false })
      expect(tree[0]?.children[0]).toMatchObject({ label: 'core', reveal: true })
      expect(tree[0]?.children[0]?.children.map((task) => task.label)).toEqual(['lint'])
    })

    it('matches the language', () => {
      expect(countProjects(toTree(single, none, { text: 'go', affectedOnly: false }))).toBe(2)
    })

    it('keeps the affected projects only', () => {
      const tree = toTree(single, none, { text: '', affectedOnly: true })
      expect(tree.map((group) => group.label)).toEqual(['libs'])
      expect(countProjects(tree)).toBe(1)
    })

    it('can be empty, and then it says why', () => {
      const filter = { text: 'zzz', affectedOnly: true }
      expect(countProjects(toTree(single, none, filter))).toBe(0)
      expect(describeFilter(filter, 0)).toBe('No project matches: affected only, “zzz”.')
    })
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

    it('hides a folder the filter leaves empty', () => {
      const both: TreeFolder[] = [{ root: '/a', name: 'a', snapshot }, { root: '/b', name: 'b', snapshot: { ...snapshot, projects: [] } }]
      expect(toTree(both, none, { text: 'web', affectedOnly: false }).map((folder) => folder.label)).toEqual(['a'])
    })

    it('keeps the ids of two folders apart', () => {
      const both: TreeFolder[] = [{ root: '/a', name: 'a', snapshot }, { root: '/b', name: 'b', snapshot }]
      const ids = toTree(both, none).flatMap((folder) => folder.children.flatMap((group) => group.children.map((project) => project.id)))
      expect(new Set(ids).size).toBe(ids.length)
    })

    it('hands each task the facts of its own folder', () => {
      const both: TreeFolder[] = [{ root: '/a', name: 'a', snapshot }, { root: '/b', name: 'b', snapshot }]
      const tree = toTree(both, (root, task) => (root === '/a' && task === 'web:build' ? { status: 'passed' } : undefined))
      const statusIn = (index: number) => tree[index]?.children[0]?.children[1]?.children[0]?.status
      expect(statusIn(0)).toBe('passed')
      expect(statusIn(1)).toBeUndefined()
    })
  })
})

describe('describeAffected', () => {
  it('counts the affected projects of all folders', () => {
    expect(describeAffected(single)).toBe('1 of 3 affected')
    expect(describeAffected([{ root: '/r', name: 'r', snapshot: { ...snapshot, affected: new Set(['web', 'core', 'api']) } }])).toBe('all 3 affected')
  })

  it('says nothing when none are affected or it is not known', () => {
    expect(describeAffected([{ root: '/r', name: 'r', snapshot: { ...snapshot, affected: new Set() } }])).toBe('')
    expect(describeAffected([{ root: '/r', name: 'r' }])).toBe('')
  })
})

describe('describeFilter', () => {
  it('says nothing when the view is not narrowed', () => {
    expect(describeFilter({ text: ' ', affectedOnly: false }, 3)).toBeUndefined()
  })

  it('says how the view is narrowed', () => {
    expect(describeFilter({ text: 'api', affectedOnly: true }, 2)).toBe('Showing affected only, “api”.')
  })
})

describe('toTree and the agent', () => {
  const missing: TreeFolder[] = [{ root: '/repo', name: 'repo', snapshot, agentMissing: true }]

  it('puts one line that adds the agent above the groups, running the install command', () => {
    const tree = toTree(missing, none)
    expect(tree.map((node) => node.label)).toEqual(['Add the MonoSmith agent', 'apps', 'libs'])
    expect(tree[0]).toMatchObject({ kind: 'action', command: 'monosmith.installAgent', root: '/repo' })
    expect(countProjects(tree)).toBe(3)
  })

  it('leaves the line out when the agent is there, not known to be missing, or the view is narrowed', () => {
    expect(toTree(single, none).map((node) => node.kind)).not.toContain('action')
    expect(toTree(missing, none, { text: 'api', affectedOnly: false }).map((node) => node.kind)).not.toContain('action')
  })

  it('puts the line under its folder when there are several', () => {
    const several: TreeFolder[] = [...missing, { root: '/other', name: 'other', snapshot }]
    const tree = toTree(several, none)
    expect(tree[0]?.children[0]?.kind).toBe('action')
    expect(tree[1]?.children[0]?.kind).toBe('group')
  })
})

describe('formatDuration', () => {
  it('reads in the unit that fits', () => {
    expect(formatDuration(850)).toBe('850ms')
    expect(formatDuration(1234)).toBe('1.2s')
    expect(formatDuration(125_000)).toBe('2m 05s')
  })
})
