import { describe, expect, it } from 'vitest'
import { isRewritable, toDependencyRows } from './updates-tree.mapper.ts'

describe('toDependencyRows', () => {
  const report = {
    updates: [
      { plugin: 'node', name: 'typescript', current: '6.0.3', latest: '7.0.2', level: 'major', sites: [{ file: 'package.json', name: 'typescript', spec: '~6.0.3', section: 'devDependencies', rewritable: true }] },
      { plugin: 'node', name: 'eslint', current: '10.12.0', latest: '10.13.0', level: 'minor', sites: [{ file: 'apps/web/package.json', name: 'eslint', spec: '^10.12.0', section: 'devDependencies', rewritable: true }, { file: 'libs/kit/package.json', name: 'eslint', spec: '^10.12.0', section: 'devDependencies', rewritable: true }] },
      { plugin: 'node', name: 'left', current: '1.0.0', latest: '1.0.1', level: 'patch', sites: [{ file: 'requirements.txt', name: 'left', spec: '==1.0.0', section: 'deps', rewritable: false }] },
    ],
    tools: [
      { plugin: 'go', name: 'Go', pinned: '1.27', latest: '1.27.3' },
      { plugin: 'python', name: 'Python', pinned: '3.12', latest: '3.15', newer: true },
    ],
  }

  it('groups the updates safe to take together, then the major ones, with the headline', () => {
    const { nodes, summary } = toDependencyRows(report)
    expect(nodes.map((node) => node.label)).toEqual([
      'Patch and minor releases · 2 · safe to take together',
      'Major releases · 1 · may break your code, take one at a time',
      'Tools',
    ])
    expect(nodes[0]?.children?.map((row) => row.label)).toEqual(['eslint', 'left'])
    expect(summary).toBe('↑ 3 with newer releases · 1 patch · 1 minor · 1 major')
  })

  it('says the versions, the level and where a package is used, with the mark of its level', () => {
    const { nodes } = toDependencyRows(report)
    expect(nodes[0]?.children?.[0]).toMatchObject({ label: 'eslint', description: '10.12.0 → 10.13.0 · minor · apps/web/package.json +1', icon: 'circle-filled', color: 'charts.blue' })
    expect(nodes[1]?.children?.[0]).toMatchObject({ label: 'typescript', description: '6.0.3 → 7.0.2 · major · package.json', icon: 'warning', color: 'charts.yellow' })
  })

  it('offers an update only for a dependency smith can rewrite, a major one marked so', () => {
    const { nodes } = toDependencyRows(report)
    expect(nodes[0]?.children?.[0]).toMatchObject({ context: 'dependency', data: 'eslint' })
    expect(nodes[0]?.children?.[1]?.context).toBeUndefined()
    expect(nodes[1]?.children?.[0]).toMatchObject({ context: 'dependency.major', data: 'typescript' })
  })

  it('puts the detail of a package under it, ending with the smith command an update runs', () => {
    const detail = toDependencyRows(report).nodes[1]?.children?.[0]?.children ?? []
    expect(detail.map((row) => row.label)).toEqual(['6.0.3 → 7.0.2', 'A major release can break your code: read its release notes first.', 'Used by', 'Declared as', 'The same as', 'Update typescript'])
    expect(detail[2]?.children).toMatchObject([{ label: 'package.json', description: 'devDependencies · ~6.0.3' }])
    expect(detail.at(-2)).toMatchObject({ description: '$ smith deps up typescript --major' })
    expect(detail.at(-1)).toMatchObject({ command: 'monosmith.dependencies.update', data: 'typescript', context: 'dependency.major.action' })
  })

  it('says where a dependency is declared in its tooltip', () => {
    expect(toDependencyRows(report).nodes[0]?.children?.[0]?.tooltip).toContain('apps/web/package.json (devDependencies, ^10.12.0)')
  })

  it('names the registry page in the tooltip and links it in the detail when smith knows it', () => {
    const withUrl = { updates: [{ plugin: 'node', name: 'zod', current: '3.22.0', latest: '3.23.1', level: 'minor', url: 'https://www.npmjs.com/package/zod' }] }
    const row = toDependencyRows(withUrl).nodes[0]?.children?.[0]
    expect(row?.tooltip).toContain('Registry: https://www.npmjs.com/package/zod')
    expect(row?.children?.find((child) => child.label === 'Registry')).toMatchObject({ command: 'monosmith.dependencies.openRegistry', data: 'https://www.npmjs.com/package/zod' })
    expect(toDependencyRows(report).nodes[0]?.children?.[0]?.tooltip).not.toContain('Registry:')
  })

  it('lists only the tools that have a newer release', () => {
    expect(toDependencyRows(report).nodes[2]?.children?.map((tool) => tool.label)).toEqual(['Python'])
  })

  it('says everything is up to date', () => {
    expect(toDependencyRows({ updates: [], tools: [{ plugin: 'go', name: 'Go', pinned: '1', latest: '1' }] })).toEqual({
      nodes: [], message: 'Every dependency is up to date.', summary: 'up to date',
    })
  })
})

describe('isRewritable', () => {
  it('needs declarations that smith can all rewrite', () => {
    expect(isRewritable({ plugin: 'node', name: 'a', current: '1', latest: '2', level: 'major' })).toBe(false)
    expect(isRewritable({ plugin: 'node', name: 'a', current: '1', latest: '2', level: 'major', sites: [{ file: 'f', name: 'a', spec: '1', section: 's' }] })).toBe(true)
  })
})
