import { describe, expect, it } from 'vitest'
import { createTaskLogStore } from './task-log.store.ts'

describe('createTaskLogStore', () => {
  it('keeps the lines of the last run of each task', () => {
    const store = createTaskLogStore()
    store.begin('api:build', '$ smith run build')
    store.add('api:build', 'compiling')
    store.begin('web:build')
    store.add('web:build', 'bundling')
    expect(store.lines('api:build')).toEqual(['$ smith run build', 'compiling'])
    expect(store.lines('web:build')).toEqual(['bundling'])
    expect(store.has('api:build')).toBe(true)
  })

  it('replaces the log when the task runs again', () => {
    const store = createTaskLogStore()
    store.begin('api:build')
    store.add('api:build', 'old')
    store.begin('api:build')
    store.add('api:build', 'new')
    expect(store.lines('api:build')).toEqual(['new'])
  })

  it('has nothing for a task that never ran, and ignores lines for it', () => {
    const store = createTaskLogStore()
    store.add('x:y', 'stray')
    expect(store.has('x:y')).toBe(false)
    expect(store.lines('x:y')).toBeUndefined()
  })

  it('keeps only the newest lines of a noisy task', () => {
    const store = createTaskLogStore(3)
    store.begin('t:t')
    for (const line of ['1', '2', '3', '4', '5']) store.add('t:t', line)
    expect(store.lines('t:t')).toEqual(['3', '4', '5'])
  })

  it('forgets everything on clear', () => {
    const store = createTaskLogStore()
    store.begin('a:b')
    store.clear()
    expect(store.has('a:b')).toBe(false)
  })
})
