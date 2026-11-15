import { describe, expect, it } from 'vitest'
import { describeOrigin } from './smith-origin.mapper.ts'

describe('describeOrigin', () => {
  it('says which smith, from where, and its path', () => {
    const lines = describeOrigin({ location: { executable: '/repo/node_modules/.bin/smith', source: 'workspace' }, version: '1.54.10', skippedUntrusted: false })
    expect(lines[0]).toBe('smith 1.54.10 from this workspace (node_modules/.bin or .smith/bin)')
    expect(lines[1]).toBe('/repo/node_modules/.bin/smith')
    expect(lines).toHaveLength(2)
  })

  it('says a workspace smith was left alone for trust', () => {
    const lines = describeOrigin({ location: { executable: '/usr/bin/smith', source: 'path' }, version: '1.56.0', skippedUntrusted: true })
    expect(lines.join('\n')).toContain('not trusted')
  })

  it('says what a terminal runs when it differs', () => {
    const lines = describeOrigin({ location: { executable: 'x', source: 'workspace' }, version: '1.54.10', skippedUntrusted: false, terminalVersion: '1.56.0' })
    expect(lines.join('\n')).toContain('A terminal runs smith 1.56.0')
  })

  it('does not repeat a terminal version that is the same', () => {
    const lines = describeOrigin({ location: { executable: 'x', source: 'setting' }, version: '1.56.0', skippedUntrusted: false, terminalVersion: '1.56.0' })
    expect(lines).toHaveLength(2)
  })
})
