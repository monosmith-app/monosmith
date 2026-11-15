import { describe, expect, it } from 'vitest'
import { createLastRunStore } from './last-run.store.ts'

describe('createLastRunStore', () => {
  it('has no run before one was asked for', () => {
    expect(createLastRunStore().last()).toBeUndefined()
  })

  it('remembers the latest run only', () => {
    const store = createLastRunStore()
    store.remember({ root: '/a', plan: { tasks: ['build'] } })
    store.remember({ root: '/b', plan: { tasks: ['test'], affected: true } })
    expect(store.last()).toEqual({ root: '/b', plan: { tasks: ['test'], affected: true } })
  })
})
