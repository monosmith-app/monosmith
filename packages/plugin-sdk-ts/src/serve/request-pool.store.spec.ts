import { describe, expect, it } from 'vitest'
import { MAX_QUEUED, RequestPool } from './request-pool.store.ts'
import type { PooledRequest } from './request-pool.store.ts'

const request = (id: number) => ({ id, method: 'discover', params: {} })

/** A pool whose runs finish when the test says so. */
function gated () {
  const started: number[] = []
  const finish = new Map<number, () => void>()
  const pool: RequestPool = new RequestPool(async (pooled: PooledRequest) => {
    started.push(pooled.request.id)
    await new Promise<void>((resolve) => { finish.set(pooled.request.id, resolve) })
    pool.release(pooled)
  })

  return { pool, started, finish }
}

describe('RequestPool', () => {
  it('runs one request at a time by default and starts the others in the order they came', async () => {
    const { pool, started, finish } = gated()
    for (const id of [1, 2, 3]) expect(pool.submit(request(id))).toBe('admitted')
    expect(started).toEqual([1])
    finish.get(1)?.()
    await new Promise(resolve => setTimeout(resolve, 0))
    expect(started).toEqual([1, 2])
    finish.get(2)?.()
    finish.get(3)?.()
  })

  it('runs as many as the limit and starts waiting ones when the limit is raised', () => {
    const { pool, started } = gated()
    pool.setLimit(2)
    for (const id of [1, 2, 3, 4]) pool.submit(request(id))
    expect(started).toEqual([1, 2])
    pool.setLimit(3)
    expect(started).toEqual([1, 2, 3])
  })

  it('refuses an id in flight and a queue that is full', () => {
    const { pool } = gated()
    pool.submit(request(1))
    expect(pool.submit(request(1))).toBe('duplicate')
    for (let id = 2; id < 2 + MAX_QUEUED; id++) expect(pool.submit(request(id))).toBe('admitted')
    expect(pool.submit(request(1_000_000))).toBe('full')
  })

  it('forgets a released request, so its id can be used again', async () => {
    const pooled: PooledRequest[] = []
    const reusable: RequestPool = new RequestPool((p) => {
      pooled.push(p); reusable.release(p)

      return Promise.resolve()
    })
    expect(reusable.submit(request(7))).toBe('admitted')
    expect(reusable.submit(request(7))).toBe('admitted')
    await reusable.idle()
    expect(pooled).toHaveLength(2)
    expect(reusable.failure).toBeUndefined()
  })

  it('cancels a running request by aborting it and a queued one by returning it', () => {
    const { pool, finish } = gated()
    pool.submit(request(1))
    pool.submit(request(2))
    expect(pool.cancel(1)).toBeUndefined()
    expect(pool.cancel(2)?.request.id).toBe(2)
    expect(pool.cancel(99)).toBeUndefined()
    expect(pool.submit(request(2))).toBe('admitted') // the cancelled one is gone from the pool
    finish.get(1)?.()
  })

  it('cancels everything and says when it is idle', async () => {
    const { pool, finish } = gated()
    pool.submit(request(1))
    pool.submit(request(2))
    pool.submit(request(3))
    expect(pool.cancelAll().map(p => p.request.id)).toEqual([2, 3])
    finish.get(1)?.()
    await pool.idle()
  })

  it('keeps the first error a run threw and goes on', async () => {
    const pool = new RequestPool(() => Promise.reject(new Error('boom')))
    pool.submit(request(1))
    pool.submit(request(2))
    await pool.idle()
    expect(pool.failure).toBeInstanceOf(Error)
  })
})
