import { describe, expect, it } from 'vitest'
import type { WorkspaceSnapshot } from '../../shared/workspace/index.ts'
import { countProjects, describeAffected, describeFilter, formatDuration, toTree, type ProjectRun, type TaskFacts, type TreeFolder, type TreeNode } from './workspace-tree.mapper.ts'

/** The task rows of a project: the rows of its Tasks section. */
const tasksOf = (project: TreeNode | undefined): TreeNode[] => project?.children.find((child) => child.kind === 'section' && child.label === 'Tasks')?.children ?? []
/** One section of a project. */
const sectionOf = (project: TreeNode | undefined, label: string): TreeNode | undefined => project?.children.find((child) => child.kind === 'section' && child.label === label)

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
/** One failed run of web:build in /repo. */
const webRuns = (root: string, project: string): ProjectRun[] => (root === '/repo' && project === 'web' ? [{ id: 'ide-1', tasks: 'build', status: 'failed', note: '2.1s · just now' }] : [])

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
    expect(tasksOf(core)).toMatchObject([
      { id: 'task:/repo:core:test', task: 'test', status: 'passed', description: '1.2s', icon: 'pass', hasLog: true, root: '/repo' },
      { id: 'task:/repo:core:lint', status: 'failed', description: 'exited with code 1', icon: 'error' },
    ])
  })

  it('does not repeat the status word of a cached task, and a task never run has the play icon', () => {
    const core = toTree(single, coreTest({ status: 'cached' }))[1]?.children[0]
    expect(tasksOf(core)[0]).toMatchObject({ status: 'cached', description: '', icon: 'history' })
    expect(tasksOf(core)[1]).toMatchObject({ description: '', icon: 'play' })
    expect(tasksOf(core)[1]?.status).toBeUndefined()
  })

  it('tells the last result, the duration and the cache hit in the tooltip', () => {
    const children = tasksOf(toTree(single, (_root, task) => (task === 'core:test' ? { status: 'cached' } : { status: 'passed', durationMs: 2500, time: '2026-10-20T06:05:59Z' }))[1]?.children[0])
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
    expect(tasksOf(core)[0]?.spoken).toBe('core test, failed in 900ms. Press Enter to run.')
    expect(tasksOf(core)[1]?.spoken).toContain('not run yet')
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
      expect(tasksOf(tree[0]?.children[0])).toHaveLength(2)
    })

    it('keeps a project for a task the text names, with that task only, open', () => {
      const tree = toTree(single, none, { text: 'lint', affectedOnly: false })
      expect(tree[0]?.children[0]).toMatchObject({ label: 'core', reveal: true })
      expect(tasksOf(tree[0]?.children[0]).map((task) => task.label)).toEqual(['lint'])
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
      const statusIn = (index: number) => tasksOf(tree[index]?.children[0]?.children[1])[0]?.status
      expect(statusIn(0)).toBe('passed')
      expect(statusIn(1)).toBeUndefined()
    })
  })
})

describe('describeAffected', () => {
  it('counts the affected projects of all folders', () => {
    expect(describeAffected(single)).toBe('1 affected')
    expect(describeAffected([{ root: '/r', name: 'r', snapshot: { ...snapshot, affected: new Set(['web', 'core', 'api']) } }])).toBe('3 affected')
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

describe('the environments of a web app', () => {
  const webApps = [{ project: 'web', default: 'prd', environments: [{ name: 'lcl', task: 'build:lcl' }, { name: 'prd', task: 'build' }] }]
  const folders: TreeFolder[] = [{ root: '/repo', name: 'repo', snapshot, webApps }]

  it('puts an Environments section under the web app with a row for each environment and the default marked', () => {
    const web = toTree(folders, none)[0]?.children[1]
    const section = web?.children.at(-1)
    expect(section).toMatchObject({ kind: 'environments', label: 'Environments', description: 'lcl, prd', project: 'web' })
    expect(section?.children).toMatchObject([
      { kind: 'environment', label: 'lcl', description: 'smith run build:lcl', project: 'web' },
      { kind: 'environment', label: 'prd', description: 'default: smith run build' },
    ])
  })

  it('adds nothing under a project that is not a web app, or when smith listed none', () => {
    const tree = toTree(folders, none)
    expect(tree[0]?.children[0]?.children.map((child) => child.kind)).not.toContain('environments')
    expect(toTree(single, none)[0]?.children[1]?.children.map((child) => child.kind)).toEqual(['section', 'section', 'section', 'section'])
  })
})

describe('the sections of a project', () => {
  it('are Tasks, Dependencies, Runs and Settings, with Tasks open', () => {
    const web = toTree(single, none)[0]?.children[1]
    expect(web?.children.map((child) => [child.label, child.open])).toEqual([['Tasks', true], ['Dependencies', false], ['Runs', false], ['Settings', false]])
    expect(sectionOf(web, 'Tasks')?.description).toBe('1')
  })

  it('say what a project uses and what uses it, and offer to change that with the smith command', () => {
    const core = toTree(single, none)[1]?.children[0]
    const rows = sectionOf(core, 'Dependencies')?.children ?? []
    expect(rows.map((row) => [row.label, row.description])).toEqual([
      ['Uses', 'none'],
      ['Used by', 'web'],
      ['Install a package', '$ smith install -p core <package>'],
      ['Unlink a library', '$ smith uninstall -p core <library>'],
    ])
    expect(rows[2]).toMatchObject({ kind: 'action', command: 'monosmith.installPackage', argument: { kind: 'project', project: 'core', root: '/repo' } })
  })

  it('list the runs of the session that ran a task of the project, each opening in Runs', () => {
    const web = toTree(single, none, undefined, webRuns)[0]?.children[1]
    expect(sectionOf(web, 'Runs')?.children).toMatchObject([{ kind: 'run', label: 'build', description: '2.1s · just now', icon: 'error', color: 'charts.red', command: 'monosmith.runs.open', argument: 'ide-1' }])
    expect(sectionOf(toTree(single, none)[0]?.children[1], 'Runs')?.children.map((row) => row.label)).toEqual(['Nothing ran yet'])
  })

  it('open the project settings with the smith command shown', () => {
    const rows = sectionOf(toTree(single, none)[0]?.children[1], 'Settings')?.children ?? []
    expect(rows).toMatchObject([{ kind: 'action', command: 'monosmith.projectSettings', description: '$ smith config set -p web <setting> <value>', argument: { kind: 'project', project: 'web' } }])
  })
})

describe('the mark of a project', () => {
  it('is the worst state of its tasks in the last run, in its colour', () => {
    const facts = new Map<string, TaskFacts>([['core:test', { status: 'passed' }], ['core:lint', { status: 'failed' }]])
    expect(toTree(single, (_root, task) => facts.get(task))[1]?.children[0]).toMatchObject({ icon: 'error', color: 'charts.red' })
    expect(toTree(single, coreTest({ status: 'passed' }))[1]?.children[0]).toMatchObject({ icon: 'pass', color: 'charts.green' })
  })

  it('is the icon of its kind before any run', () => {
    expect(toTree(single, none)[1]?.children[0]?.color).toBeUndefined()
  })
})

describe('the filter words of the terminal', () => {
  it('read kind:, lang: and changed', () => {
    expect(countProjects(toTree(single, none, { text: 'kind:app', affectedOnly: false }))).toBe(2)
    expect(countProjects(toTree(single, none, { text: 'lang:go kind:lib', affectedOnly: false }))).toBe(1)
    expect(countProjects(toTree(single, none, { text: 'changed', affectedOnly: false }))).toBe(1)
  })
})
