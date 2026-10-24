import type { RefreshScope } from './refresh-scope.contract.ts'

/** The clock the coalescer waits on; the real one unless a test brings its own. */
export interface RefreshTimers {
  start: (callback: () => void, delayMs: number) => unknown
  stop:  (handle: unknown) => void
}

const realTimers: RefreshTimers = {
  start: (callback, delayMs) => setTimeout(callback, delayMs),
  stop:  (handle) => { clearTimeout(handle as ReturnType<typeof setTimeout>) },
}

/** Turns a burst of requests into one refresh. */
export interface RefreshCoalescer {
  /** Asks for a refresh; the wait starts over, and the widest scope asked wins. */
  request: (scope: RefreshScope) => void
  /** Drops a refresh that is waiting. */
  cancel:  () => void
}

/**
 * Waits for `delayMs` of quiet after the last request, then fires once with
 * the widest scope asked for in the burst. Saving twenty files or switching
 * branch costs one refresh, and `cancel` makes a waiting one never happen.
 */
export function createRefreshCoalescer (
  delayMs: number,
  fire: (scope: RefreshScope) => void,
  timers: RefreshTimers = realTimers,
): RefreshCoalescer {
  let pending: RefreshScope | undefined
  let handle: unknown

  const cancel = (): void => {
    if (handle !== undefined) timers.stop(handle)
    handle = undefined
    pending = undefined
  }

  return {
    request: (scope) => {
      pending = pending === 'full' || scope === 'full' ? 'full' : 'affected'
      if (handle !== undefined) timers.stop(handle)
      handle = timers.start(() => {
        const scopeToFire = pending
        handle = undefined
        pending = undefined
        if (scopeToFire !== undefined) fire(scopeToFire)
      }, delayMs)
    },
    cancel,
  }
}
