import { describe, expect, it } from 'vitest'
import { toDependencyRows } from './updates-tree.mapper.ts'

describe('toDependencyRows', () => {
  const report = {
    updates: [
      { plugin: 'node', name: 'typescript', current: '6.0.3', latest: '7.0.2', level: 'major', sites: [{ file: 'package.json', name: 'typescript', spec: '~6.0.3', section: 'devDependencies', rewritable: true }] },
      { plugin: 'node', name: 'eslint', current: '10.12.0', latest: '10.13.0', level: 'minor', sites: [{ file: 'package.json', name: 'eslint', spec: '^10.12.0', section: 'devDependencies', rewritable: true }] },
      { plugin: 'node', name: 'left', current: '1.0.0', latest: '1.0.1', level: 'patch', sites: [{ file: 'requirements.txt', name: 'left', spec: '==1.0.0', section: 'deps', rewritable: false }] },
    ],
    tools: [
      { plugin: 'go', name: 'Go', pinned: '1.27', latest: '1.27.3' },
      { plugin: 'python', name: 'Python', pinned: '3.12', latest: '3.15', newer: true },
    ],
  }

  it('groups the updates patch, minor, major and counts them', () => {
    const { nodes, summary } = toDependencyRows(report)
    expect(nodes.map((node) => `${node.label} ${node.description}`)).toEqual(['Patch 1', 'Minor 1', 'Major 1', 'Tools 1'])
    expect(nodes[2]?.children?.[0]).toMatchObject({ label: 'typescript', description: '→ 7.0.2', color: 'charts.red' })
    expect(summary).toBe('3 updates')
  })

  it('offers an update only for a dependency smith can rewrite', () => {
    const { nodes } = toDependencyRows(report)
    expect(nodes[1]?.children?.[0]).toMatchObject({ context: 'dependency', data: 'eslint' })
    expect(nodes[0]?.children?.[0]?.context).toBeUndefined()
  })

  it('says where a dependency is declared in its tooltip', () => {
    expect(toDependencyRows(report).nodes[1]?.children?.[0]?.tooltip).toContain('package.json (devDependencies, ^10.12.0)')
  })

  it('lists only the tools that have a newer release', () => {
    expect(toDependencyRows(report).nodes[3]?.children?.map((tool) => tool.label)).toEqual(['Python'])
  })

  it('says everything is up to date', () => {
    expect(toDependencyRows({ updates: [], tools: [{ plugin: 'go', name: 'Go', pinned: '1', latest: '1' }] })).toEqual({
      nodes: [], message: 'Every dependency is up to date.', summary: 'up to date',
    })
  })
})
