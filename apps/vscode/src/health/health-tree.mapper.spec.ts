import { describe, expect, it } from 'vitest'
import { toHealthRows } from './health-tree.mapper.ts'

describe('toHealthRows', () => {
  const report = {
    findings: [
      { check: 'cache ignored', ok: false, detail: '.smith/ is not in .gitignore', remedy: 'add `.smith/` to .gitignore' },
      { check: 'tracked files', ok: false, warning: true, detail: 'generated files are tracked' },
      { check: 'has commits', ok: true, remedy: 'commit the workspace' },
    ],
  }

  it('puts the failed checks first with the problem and the remedy under them', () => {
    const { nodes, summary } = toHealthRows(report)
    expect(nodes.map((node) => node.label)).toEqual(['cache ignored', 'tracked files', 'Passed'])
    expect(nodes[0]).toMatchObject({ icon: 'error', color: 'charts.red', expanded: true })
    expect(nodes[0]?.children?.map((child) => child.label)).toEqual(['.smith/ is not in .gitignore', 'add `.smith/` to .gitignore'])
    expect(nodes[1]).toMatchObject({ icon: 'warning' })
    expect(summary).toBe('1 failed, 1 warning')
  })

  it('does not show the remedy of a check that passed as a problem', () => {
    const passed = toHealthRows(report).nodes[2]
    expect(passed?.description).toBe('1')
    expect(passed?.children?.[0]).toMatchObject({ label: 'has commits' })
    expect(passed?.children?.[0]?.children).toBeUndefined()
  })

  it('lists the fixes smith proposes', () => {
    const { nodes } = toHealthRows(report, { titles: ['ignore the cache directory'], text: '● ignore the cache directory\n+.smith/' })
    const fixes = nodes.find((node) => node.id === 'health:fixes')
    expect(fixes?.children).toMatchObject([{ label: 'ignore the cache directory', icon: 'lightbulb' }])
  })

  it('says all passed when nothing failed', () => {
    const result = toHealthRows({ findings: [{ check: 'a', ok: true }, { check: 'b', ok: true }] })
    expect(result.summary).toBe('all passed')
    expect(result.nodes).toMatchObject([{ label: 'All 2 checks passed', icon: 'pass' }])
  })

  it('says so when smith reports no checks', () => {
    expect(toHealthRows({})).toEqual({ nodes: [], message: 'smith reported no checks.' })
  })
})
