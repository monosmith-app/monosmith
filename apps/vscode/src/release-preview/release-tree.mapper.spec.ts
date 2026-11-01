import { describe, expect, it } from 'vitest'
import { toReleaseRows } from './release-tree.mapper.ts'

describe('toReleaseRows', () => {
  const preview = {
    dryRun:  true,
    results: [
      { project: 'sdk', current: '1.2.0', next: '1.3.0', level: 'minor', tag: 'sdk/v1.3.0', reason: 'feat: add a client', commits: ['feat(sdk): add a client', 'fix(sdk): typo'] },
      { project: 'tools', next: '0.1.0', level: 'major', reason: 'first release' },
    ],
    unchanged: ['core'],
  }

  it('shows each package with the version it moves to and the level', () => {
    const { nodes, summary } = toReleaseRows(preview)
    expect(nodes[0]).toMatchObject({ label: 'sdk', description: '1.2.0 → 1.3.0 · minor', color: 'charts.yellow' })
    expect(nodes[1]).toMatchObject({ label: 'tools', description: 'new → 0.1.0 · major', color: 'charts.red' })
    expect(summary).toBe('2 to release')
  })

  it('puts the reason, the tag and the commits under a package', () => {
    const children = toReleaseRows(preview).nodes[0]?.children ?? []
    expect(children.map((child) => child.label)).toEqual(['feat: add a client', 'sdk/v1.3.0', 'feat(sdk): add a client', 'fix(sdk): typo'])
  })

  it('lists the packages with nothing to release, closed', () => {
    const last = toReleaseRows(preview).nodes.at(-1)
    expect(last).toMatchObject({ label: 'Unchanged', description: '1', children: [{ label: 'core' }] })
    expect(last?.expanded).toBeUndefined()
  })

  it('says so when there is nothing to release', () => {
    expect(toReleaseRows({ dryRun: true, results: [], unchanged: [] })).toEqual({ nodes: [], message: 'Nothing to release.', summary: 'nothing to release' })
  })

  it('puts the release channel and its notice first, and in the summary', () => {
    const { nodes, summary } = toReleaseRows({ ...preview, channel: 'next' })
    expect(nodes[0]).toMatchObject({ label: 'Release channel next', description: 'prerelease versions' })
    expect(nodes[0]?.children?.[0]?.label).toBe('Release channel `next`: prerelease versions, published under the `next` tag; the stable version stays the latest.')
    expect(nodes[1]).toMatchObject({ label: 'sdk' })
    expect(summary).toBe('2 to release · channel next')
  })

  it('shows no channel row on the stable line', () => {
    expect(toReleaseRows({ ...preview, channel: '' }).nodes[0]).toMatchObject({ label: 'sdk' })
  })

  it('keeps the channel in the summary when nothing releases', () => {
    expect(toReleaseRows({ channel: 'beta', results: [] }).summary).toBe('nothing to release · channel beta')
  })

  it('shows an error the plan stopped with', () => {
    const { nodes, message } = toReleaseRows({ error: 'not on the release branch' })
    expect(nodes[0]).toMatchObject({ label: 'not on the release branch', icon: 'error' })
    expect(message).toBeUndefined()
  })
})
