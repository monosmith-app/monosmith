import { describe, expect, it } from 'vitest'
import type { HomeFacts } from './home-facts.contract.ts'
import { toHomeRows } from './home-rows.mapper.ts'

const facts: HomeFacts = {
  projectsLoaded: true,
  projects:       3,
  languages:      ['Go', 'TypeScript', 'Go'],
  affected:       2,
  health:         { failed: 0, warnings: 1, passed: 9, fixable: 0, problems: [], warningNotes: ['tsconfig drifts'] },
  last:           { command: 'smith run build', running: false, ok: true, stopped: false, took: '3.0s', ago: '3 min ago', passed: 2, cached: 1, failed: 0 },
  release:        { packages: 2, name: 'kit', from: '1.3.2', to: '1.4.0', commits: 0, major: 0, minor: 1, patch: 1 },
  releaseKnown:   true,
  outdated:       0,
}

describe('toHomeRows', () => {
  it('starts with the four boxes of the terminal, each opening its view', () => {
    const { nodes, summary } = toHomeRows(facts)
    expect(summary).toBe('3 projects (Go, TypeScript) · 2 changed · last run 3 min ago')
    expect(nodes.slice(0, 4).map((node) => [node.label, node.description, node.command])).toEqual([
      ['Changed', '2 projects', 'monosmith.projects.focus'],
      ['Last run', '2 passed · 1 cached', 'monosmith.runs.open'],
      ['Health', '1 warning · 0 failed checks', 'monosmith.health.focus'],
      ['Release', '2 ready · 1 minor · 1 patch', 'monosmith.release.focus'],
    ])
    expect(nodes[2]).toMatchObject({ icon: 'warning', color: 'charts.yellow' })
  })

  it('lists Needs you now, each item running its first button and showing its buttons with their commands', () => {
    const queue = toHomeRows(facts).nodes[4]
    expect(queue).toMatchObject({ label: 'Needs you now', description: '3 items', expanded: true })
    const [warning] = queue?.children ?? []
    expect(warning).toMatchObject({ label: '1 warning: tsconfig drifts', description: 'Warning', command: 'monosmith.health.focus', expanded: true })
    expect(queue?.children?.[1]).toMatchObject({ label: '2 projects are affected by your changes', expanded: false })
    expect(queue?.children?.[1]?.children).toMatchObject([{ label: 'Test what changed', description: '$ smith run lint test --affected', command: 'monosmith.testChanged' }])
  })

  it('ends with Start something and the smith command of each action', () => {
    const start = toHomeRows(facts).nodes.at(-1)
    expect(start?.label).toBe('Start something')
    expect(start?.children?.map((row) => [row.label, row.description])).toEqual([
      ['Build all', '$ smith run build'],
      ['Lint all', '$ smith run lint'],
      ['Run a task…', '$ smith run <task> -p <project>'],
      ['Add a project', '$ smith add'],
      ['Compose a stack', '$ smith compose'],
      ['Settings', undefined],
    ])
  })

  it('leaves Start something out of an empty workspace, whose one item offers the same', () => {
    expect(toHomeRows({ ...facts, projects: 0 }).nodes.map((node) => node.label)).not.toContain('Start something')
  })

  it('says what each box waits for, or that it has nothing', () => {
    const { health: _health, last: _last, release: _release, ...bare } = facts
    const boxes = toHomeRows({ ...bare, affected: 0, releaseKnown: false }).nodes.slice(0, 4).map((node) => node.description)
    expect(boxes).toEqual(['Nothing changed', 'Nothing ran yet', 'Loading…', 'Loading…'])
    expect(toHomeRows({ ...bare, releaseKnown: true }).nodes[3]?.description).toBe('Nothing ready')
  })

  it('tells a failed run with its count', () => {
    const failed = toHomeRows({ ...facts, last: { ...facts.last!, ok: false, failed: 2 } }).nodes[1]
    expect(failed).toMatchObject({ description: '2 failed · 2 passed · 1 cached', icon: 'error' })
  })
})
