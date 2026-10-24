import { describe, expect, it } from 'vitest'
import { describeStatus, type StatusInput } from './status-text.mapper.ts'

const ready = (over: Partial<StatusInput> = {}): StatusInput => ({
  state:    'ready',
  version:  '1.52.1',
  affected: 3,
  running:  false,
  failed:   0,
  origin:   ['smith 1.52.1 from the PATH', '/usr/bin/smith'],
  ...over,
})

describe('describeStatus', () => {
  it('shows the version and the affected count', () => {
    expect(describeStatus(ready()).text).toBe('$(tools) smith 1.52.1 · 3 affected')
  })

  it('leaves the count out when it is not known', () => {
    expect(describeStatus(ready({ affected: undefined })).text).toBe('$(tools) smith 1.52.1')
  })

  it('spins while a run goes', () => {
    expect(describeStatus(ready({ running: true })).text).toContain('$(sync~spin)')
  })

  it('tells which smith in the tooltip, and what a click does', () => {
    const { tooltip } = describeStatus(ready())
    expect(tooltip).toContain('/usr/bin/smith')
    expect(tooltip).toContain('Click for a menu')
  })

  it('counts failed tasks and turns to a warning', () => {
    const status = describeStatus(ready({ failed: 2 }))
    expect(status.text).toBe('$(warning) smith 1.52.1 · 3 affected · 2 failed')
    expect(status.warning).toBe(true)
  })

  it('is not a warning when nothing failed', () => {
    expect(describeStatus(ready()).warning).toBe(false)
  })

  it('says why smith cannot be used', () => {
    expect(describeStatus({ ...ready(), state: 'missing' }).tooltip).toContain('not found')
    expect(describeStatus({ ...ready(), state: 'untrusted' }).tooltip).toContain('not trusted')
    expect(describeStatus({ ...ready(), state: 'failed' }).tooltip).toContain('could not start')
    expect(describeStatus({ ...ready(), state: 'missing' }).warning).toBe(true)
  })

  it('does not warn while starting', () => {
    expect(describeStatus({ ...ready(), state: 'starting' }).warning).toBe(false)
  })
})
