import { describe, expect, it } from 'vitest'
import { publishes } from './publishes.policy.ts'

describe('publishes', () => {
  it('is true for a package, unless the registry is none', () => {
    expect(publishes({ kind: 'package' })).toBe(true)
    expect(publishes({ kind: 'package', registries: { publish: { name: 'r', auth: 'token' } } })).toBe(true)
    expect(publishes({ kind: 'package', registries: { noPublish: true } })).toBe(false)
    expect(publishes({ kind: 'lib' })).toBe(false)
    expect(publishes({ kind: 'app' })).toBe(false)
  })
})
