import { describe, expect, it } from 'vitest'
import type { DeriveReport } from './derive-report.contract.ts'
import { addDeriveRows, contextFor, fixFor } from './derive-tree.mapper.ts'

const health = { nodes: [{ id: 'health:ok', label: 'All 3 checks passed' }], summary: 'all passed' }
const pending: DeriveReport = {
  state: 'pending',
  total: 4,
  files: [
    { path: 'packages/web/tsconfig.lib.json', status: 'drift', renderer: 'tsconfig-references', detail: 'edited by hand' },
    { path: 'packages/api/jest.config.ts', status: 'stale', renderer: 'jest-display-name', detail: '' },
  ],
}

describe('addDeriveRows', () => {
  it('adds nothing when the feature is off', () => {
    expect(addDeriveRows(health, { state: 'off', files: [], total: 0 })).toBe(health)
  })

  it('counts the files up to date', () => {
    const { nodes, summary } = addDeriveRows(health, { state: 'current', files: [], total: 7 })
    expect(nodes.at(-1)).toMatchObject({ label: 'Derived files', description: '7 up to date', icon: 'pass' })
    expect(summary).toBe('all passed')
  })

  it('lists each file that needs action with its state, its menu context, its path and the fix', () => {
    const { nodes, summary } = addDeriveRows(health, pending)
    const row = nodes.at(-1)
    expect(row).toMatchObject({ label: 'Derived files', description: '2 need action', expanded: true })
    expect(row?.children?.[0]).toMatchObject({ label: 'packages/web/tsconfig.lib.json', description: 'drift · tsconfig-references', context: 'derived.held', data: 'packages/web/tsconfig.lib.json' })
    expect(row?.children?.[0]?.children?.map((child) => child.label)).toEqual(['edited by hand', `Fix: ${fixFor('drift')}`])
    expect(row?.children?.[1]).toMatchObject({ context: 'derived.file' })
    expect(row?.children?.[1]?.children?.map((child) => child.label)).toEqual([`Fix: ${fixFor('stale')}`])
    expect(summary).toBe('2 derived files need action')
  })

  it('adds the count to the summary of the doctor', () => {
    expect(addDeriveRows({ ...health, summary: '1 failed' }, pending).summary).toBe('1 failed, 2 derived files need action')
  })

  it('shows what smith said when it could not check', () => {
    const { nodes } = addDeriveRows(health, { state: 'failed', files: [], total: 0, message: 'smith.toml is not valid Try: fix line 3' })
    expect(nodes.at(-1)).toMatchObject({ label: 'Derived files: smith.toml is not valid Try: fix line 3', icon: 'error' })
  })
})

describe('contextFor', () => {
  it('holds back drift, diverged and broken files', () => {
    expect(contextFor('drift')).toBe('derived.held')
    expect(contextFor('missing')).toBe('derived.file')
  })
})
