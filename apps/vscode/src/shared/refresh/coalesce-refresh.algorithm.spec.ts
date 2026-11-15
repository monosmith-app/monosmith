import { describe, expect, it } from 'vitest'
import { createRefreshCoalescer, type RefreshTimers } from './coalesce-refresh.algorithm.ts'
import type { RefreshScope } from './refresh-scope.contract.ts'

/** A clock the test moves by hand. */
function fakeClock () {
  let now = 0
  let next = 1
  const waiting = new Map<number, { at: number, callback: () => void }>()
  const timers: RefreshTimers = {
    start: (callback, delayMs) => {
      const id = next++
      waiting.set(id, { at: now + delayMs, callback })

      return id
    },
    stop: (handle) => { waiting.delete(handle as number) },
  }
  const advance = (ms: number): void => {
    now += ms
    for (const [id, timer] of waiting) {
      if (!(timer.at <= now)) {
        continue
      }

      waiting.delete(id)
      timer.callback()
    }
  }

  return { timers, advance, pending: () => waiting.size }
}

describe('createRefreshCoalescer', () => {
  it('fires once after a quiet period, however many requests came', () => {
    const clock = fakeClock()
    const fired: RefreshScope[] = []
    const coalescer = createRefreshCoalescer(500, (scope) => { fired.push(scope) }, clock.timers)
    coalescer.request('affected')
    clock.advance(300)
    coalescer.request('affected')
    clock.advance(300)
    expect(fired).toEqual([])
    clock.advance(200)
    expect(fired).toEqual(['affected'])
  })

  it('fires with the widest scope of the burst', () => {
    const clock = fakeClock()
    const fired: RefreshScope[] = []
    const coalescer = createRefreshCoalescer(100, (scope) => { fired.push(scope) }, clock.timers)
    coalescer.request('full')
    coalescer.request('affected')
    clock.advance(100)
    expect(fired).toEqual(['full'])
  })

  it('starts narrow again after firing', () => {
    const clock = fakeClock()
    const fired: RefreshScope[] = []
    const coalescer = createRefreshCoalescer(100, (scope) => { fired.push(scope) }, clock.timers)
    coalescer.request('full')
    clock.advance(100)
    coalescer.request('affected')
    clock.advance(100)
    expect(fired).toEqual(['full', 'affected'])
  })

  it('never fires a request that was cancelled', () => {
    const clock = fakeClock()
    const fired: RefreshScope[] = []
    const coalescer = createRefreshCoalescer(100, (scope) => { fired.push(scope) }, clock.timers)
    coalescer.request('full')
    coalescer.cancel()
    clock.advance(1000)
    expect(fired).toEqual([])
    expect(clock.pending()).toBe(0)
  })
})
