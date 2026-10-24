import { describe, expect, it } from 'vitest'
import type { RunEvent } from './run-event.contract.ts'
import { createTaskStatusStore } from './task-status.store.ts'

const at = (event: Partial<RunEvent> & { event: string }): RunEvent => ({ run: 'r', seq: 1, time: '', ...event })

describe('failedCount', () => {
  it('counts the tasks left failed', () => {
    const store = createTaskStatusStore()
    store.apply(at({ event: 'task.finished', task: 'api:build', status: 'failed' }))
    store.apply(at({ event: 'task.finished', task: 'web:build', status: 'passed' }))
    expect(store.failedCount()).toBe(1)
    store.apply(at({ event: 'task.finished', task: 'api:build', status: 'passed' }))
    expect(store.failedCount()).toBe(0)
  })
})

describe('createTaskStatusStore', () => {
  it('follows a task from queued to its end', () => {
    const store = createTaskStatusStore()
    store.apply(at({ event: 'run.started', tasks: ['api:build', 'api:test'] }))
    expect(store.statusOf('api:test')).toBe('queued')
    store.apply(at({ event: 'task.started', task: 'api:build' }))
    expect(store.statusOf('api:build')).toBe('running')
    store.apply(at({ event: 'task.finished', task: 'api:build', status: 'cached' }))
    expect(store.statusOf('api:build')).toBe('cached')
  })

  it('knows nothing of a task no run touched', () => {
    expect(createTaskStatusStore().statusOf('web:lint')).toBeUndefined()
  })

  it('tells its listeners when a status changed, not for output lines', () => {
    const store = createTaskStatusStore()
    let changes = 0
    const stop = store.onChange(() => { changes++ })
    store.apply(at({ event: 'task.started', task: 'api:build' }))
    store.apply(at({ event: 'task.output', task: 'api:build', line: 'compiling' }))
    expect(changes).toBe(1)
    stop()
    store.apply(at({ event: 'task.finished', task: 'api:build', status: 'passed' }))
    expect(changes).toBe(1)
  })
})

describe('resultOf', () => {
  it('keeps how long a task ran, why it failed and when', () => {
    const store = createTaskStatusStore()
    store.apply(at({ event: 'task.finished', task: 'api:build', status: 'failed', durationMs: 1200, reason: 'exited with code 1', time: '2026-10-20T06:05:59Z' }))
    expect(store.resultOf('api:build')).toEqual({ status: 'failed', durationMs: 1200, reason: 'exited with code 1', time: '2026-10-20T06:05:59Z' })
    expect(store.resultOf('web:build')).toBeUndefined()
  })
})
