import { describe, expect, it } from 'vitest'
import { auditBlocks, auditEnabled } from './audit-settings.policy.ts'

describe('audit settings', () => {
  it('enables only the listed policies and blocks unless a policy only warns', () => {
    const audit = { policies: ['node', 'go'], warn: ['go'] }
    expect(auditEnabled(audit, 'node')).toBe(true)
    expect(auditEnabled(audit, 'rust')).toBe(false)
    expect(auditEnabled(undefined, 'node')).toBe(false)
    expect(auditBlocks(audit, 'node')).toBe(true)
    expect(auditBlocks(audit, 'go')).toBe(false)
    expect(auditBlocks(undefined, 'go')).toBe(true)
  })
})
