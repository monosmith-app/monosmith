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

describe('describeStatus for a folder that can be converted', () => {
  it('says the folder can be converted and does not warn', () => {
    const status = describeStatus(ready({ state: 'adoptable' }))
    expect(status.tooltip).toContain('can be converted')
    expect(status.warning).toBe(false)
  })
})

describe('describeStatus for a folder with no smith.toml and nothing to convert', () => {
  it('says an empty folder is empty and offers a new workspace', () => {
    const status = describeStatus(ready({ state: 'empty' }))
    expect(status.tooltip).toContain('this folder is empty')
    expect(status.tooltip).toContain('create a new workspace')
    expect(status.warning).toBe(false)
  })

  it('says a folder with files is not a workspace yet and offers to bring it in', () => {
    const status = describeStatus(ready({ state: 'unrecognised' }))
    expect(status.tooltip).toContain('not a MonoSmith workspace yet')
    expect(status.tooltip).toContain('bring this folder in')
  })

  it('says no folder is open', () => {
    expect(describeStatus(ready({ state: 'nofolder' })).tooltip).toContain('no folder is open')
  })
})
