import { describe, expect, it } from 'vitest'
import { normalizeSettings } from './normalize-settings.mapper.ts'

describe('normalizeSettings', () => {
  it('has the defaults when nothing was written', () => {
    expect(normalizeSettings({})).toEqual({
      refreshAuto: true, refreshDebounceMs: 1500, parallel: 0, problems: true, bail: false, noCache: false,
    })
  })

  it('takes what was written', () => {
    expect(normalizeSettings({ refreshAuto: false, refreshDebounceMs: 400, parallel: 3, problems: false, bail: true, noCache: true })).toEqual({
      refreshAuto: false, refreshDebounceMs: 400, parallel: 3, problems: false, bail: true, noCache: true,
    })
  })

  it('holds the wait and the parallelism between their limits', () => {
    expect(normalizeSettings({ refreshDebounceMs: 5 }).refreshDebounceMs).toBe(100)
    expect(normalizeSettings({ refreshDebounceMs: 9e9 }).refreshDebounceMs).toBe(30_000)
    expect(normalizeSettings({ parallel: -2 }).parallel).toBe(0)
    expect(normalizeSettings({ parallel: 2.7 }).parallel).toBe(2)
  })

  it('falls back to the default for a value of the wrong type', () => {
    expect(normalizeSettings({ refreshAuto: 'yes', refreshDebounceMs: '500', parallel: NaN })).toMatchObject({
      refreshAuto: true, refreshDebounceMs: 1500, parallel: 0,
    })
  })
})
