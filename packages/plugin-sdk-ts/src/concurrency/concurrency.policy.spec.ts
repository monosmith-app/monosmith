import { describe, expect, it } from 'vitest'
import { CONCURRENCY_UNLIMITED, UNLIMITED_IN_FLIGHT, UNLIMITED_WORKERS, concurrencyLimit, inFlightLimit, isValidConcurrency, workerLimit } from './concurrency.policy.ts'

describe('concurrency', () => {
  it('is one request at a time unless the plugin declares more', () => {
    for (const declared of [undefined, 0, 1, -2, -100, 1.5]) expect(concurrencyLimit(declared)).toEqual({ limit: 1, unlimited: false })
    expect(concurrencyLimit(4)).toEqual({ limit: 4, unlimited: false })
  })

  it('leaves the number to the host for -1, with bounds on both sides', () => {
    expect(concurrencyLimit(CONCURRENCY_UNLIMITED).unlimited).toBe(true)
    expect(inFlightLimit(CONCURRENCY_UNLIMITED)).toBe(UNLIMITED_IN_FLIGHT)
    expect(workerLimit(CONCURRENCY_UNLIMITED)).toBe(UNLIMITED_WORKERS)
    expect(inFlightLimit(3)).toBe(3)
    expect(workerLimit(undefined)).toBe(1)
  })

  it('defines absent, a count and unlimited, and nothing below', () => {
    expect([-2, -1, 0, 1, 64].map(declared => isValidConcurrency(declared))).toEqual([false, true, true, true, true])
    expect(isValidConcurrency(1.5)).toBe(false)
  })
})
