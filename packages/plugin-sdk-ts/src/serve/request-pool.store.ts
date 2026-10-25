import type { ReceivedRequest } from '../codec/index.ts'

/** The requests waiting for a worker are limited to this many; the host never has more in flight than the plugin declared, so more means a host that breaks the protocol. */
export const MAX_QUEUED = 1024

/** One request from the moment it is read until it is answered. */
export interface PooledRequest {
  readonly request:    ReceivedRequest & { id: number }
  /** Aborted when the host cancels the request (or the plugin is shutting down). */
  readonly controller: AbortController
  /** Whether the host asked for the request to stop. */
  cancelled:           boolean
}

/** What {@link RequestPool.submit} decided. */
export type Admission = 'admitted' | 'duplicate' | 'full'

/**
 * The state of the serve loop: the requests in flight, the ones waiting for a
 * worker, and a bounded number of workers running them. Requests start in the
 * order they arrive and may finish in any order. The Go SDK's request pool,
 * with the same rules.
 */
export class RequestPool {
  #limit = 1
  #running = 0
  readonly #queue: PooledRequest[] = []
  readonly #inflight = new Map<number, PooledRequest>()
  readonly #idle:  Array<() => void> = []
  readonly #run:   (pooled: PooledRequest) => Promise<void>
  #failure:        unknown

  /** `run` answers one request; if it throws, the first error is kept in {@link failure} and the pool goes on. */
  constructor (run: (pooled: PooledRequest) => Promise<void>) {
    this.#run = run
  }

  #start (first: PooledRequest): void {
    this.#running++
    void this.#work(first)
  }

  async #work (first: PooledRequest): Promise<void> {
    for (let pooled: PooledRequest | undefined = first; pooled !== undefined; pooled = this.#next()) {
      try {
        await this.#run(pooled)
      } catch (error) {
        this.#failure ??= error
      }
    }
  }

  #next (): PooledRequest | undefined {
    if (this.#queue.length > 0 && this.#running <= this.#limit) return this.#queue.shift()
    this.#running--
    if (this.#running === 0) for (const resolve of this.#idle.splice(0)) resolve()

    return undefined
  }

  /** The first error a run threw, if any. */
  get failure (): unknown {
    return this.#failure
  }

  /** How many workers may run at once. Raising it starts waiting requests; lowering it lets running ones finish. */
  setLimit (limit: number): void {
    this.#limit = Math.max(1, limit)
    while (this.#running < this.#limit && this.#queue.length > 0) this.#start(this.#queue.shift() as PooledRequest)
  }

  /** Takes a request: it runs at once on a free worker, or waits its turn. */
  submit (request: ReceivedRequest & { id: number }): Admission {
    if (this.#inflight.has(request.id)) return 'duplicate'
    if (this.#running >= this.#limit && this.#queue.length >= MAX_QUEUED) return 'full'
    const pooled: PooledRequest = { request, controller: new AbortController(), cancelled: false }
    this.#inflight.set(request.id, pooled)
    if (this.#running >= this.#limit) this.#queue.push(pooled)
    else this.#start(pooled)

    return 'admitted'
  }

  /**
   * Forgets an answered request; call it before the answer is written, so a
   * host that reuses the id right away is not refused as a duplicate.
   */
  release (pooled: PooledRequest): void {
    this.#inflight.delete(pooled.request.id)
  }

  /**
   * Stops the request with an id. One that had not started is taken out of
   * the queue and returned, because no worker will ever answer it. An unknown
   * id is ignored.
   */
  cancel (id: number): PooledRequest | undefined {
    const pooled = this.#inflight.get(id)
    if (pooled === undefined) return undefined
    pooled.cancelled = true
    pooled.controller.abort()
    const at = this.#queue.indexOf(pooled)
    if (at === -1) return undefined
    this.#queue.splice(at, 1)
    this.#inflight.delete(id)

    return pooled
  }

  /** Stops every request. The ones that had not started are returned, to be answered. */
  cancelAll (): PooledRequest[] {
    for (const pooled of this.#inflight.values()) {
      pooled.cancelled = true
      pooled.controller.abort()
    }
    const waiting = this.#queue.splice(0)
    for (const pooled of waiting) this.#inflight.delete(pooled.request.id)

    return waiting
  }

  /** Resolves once no worker is running. */
  idle (): Promise<void> {
    if (this.#running === 0) return Promise.resolve()

    return new Promise<void>((resolve) => { this.#idle.push(resolve) })
  }
}
