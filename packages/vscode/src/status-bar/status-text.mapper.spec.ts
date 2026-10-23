import { describe, expect, it } from 'vitest'
import { describeStatus } from './status-text.mapper.ts'

describe('describeStatus', () => {
  it('shows the version and the affected count', () => {
    expect(describeStatus('1.52.1', 3, false).text).toBe('$(tools) smith 1.52.1 · 3 affected')
  })

  it('leaves the count out when it is not known', () => {
    expect(describeStatus('1.52.1', undefined, false).text).toBe('$(tools) smith 1.52.1')
  })

  it('spins while a run goes', () => {
    expect(describeStatus('1.52.1', 0, true).text).toContain('$(sync~spin)')
  })

  it('says smith is missing when there is no version', () => {
    expect(describeStatus(undefined, undefined, false).tooltip).toContain('not found')
  })
})
