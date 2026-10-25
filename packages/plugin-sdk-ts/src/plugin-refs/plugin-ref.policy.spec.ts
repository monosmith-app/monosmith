import { describe, expect, it } from 'vitest'
import { conflictsWith, describeRef, isEmptyRef, requiresLanguage } from './plugin-ref.policy.ts'

describe('plugin references', () => {
  it('builds the two common ones', () => {
    expect(requiresLanguage('node')).toEqual({ level: 'language', name: 'node' })
    expect(conflictsWith('x')).toEqual({ name: 'x' })
  })

  it('knows a reference that names nothing', () => {
    expect(isEmptyRef({})).toBe(true)
    expect(isEmptyRef({ level: 'language' })).toBe(false)
    expect(isEmptyRef({ slot: 'test' })).toBe(false)
  })

  it('names a reference as an error message does', () => {
    expect(describeRef({})).toBe('plugin')
    expect(describeRef(requiresLanguage('node'))).toBe('language plugin node')
    expect(describeRef({ slot: 'test' })).toBe('plugin providing test')
    expect(describeRef({ level: 'toolchain', name: 'jest', slot: 'test' })).toBe('toolchain plugin jest providing test')
  })
})
