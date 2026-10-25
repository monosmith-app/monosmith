import { describe, expect, it } from 'vitest'
import { setToolchain } from './set-toolchain.algorithm.ts'

describe('setToolchain', () => {
  it('probes once and records the toolchain on every project', () => {
    let probes = 0
    const result = setToolchain({ projects: [{ name: 'a', root: 'a', language: 'l', kind: 'lib' }, { name: 'b', root: 'b', language: 'l', kind: 'lib' }] }, () => {
      probes++

      return 'node 24.0.0'
    })
    expect(probes).toBe(1)
    expect(result.projects.map(p => p.toolchain)).toEqual(['node 24.0.0', 'node 24.0.0'])
  })

  it('does not probe when there is no project', () => {
    let probes = 0
    setToolchain({ projects: [] }, () => {
      probes++

      return ''
    })
    expect(probes).toBe(0)
  })
})
