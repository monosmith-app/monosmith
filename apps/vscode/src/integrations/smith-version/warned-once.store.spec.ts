import { describe, expect, it } from 'vitest'
import { createWarnedOnceStore } from './warned-once.store.ts'

describe('createWarnedOnceStore', () => {
  it('is true the first time for a key and false after', () => {
    const store = createWarnedOnceStore()
    expect(store.firstTime('a')).toBe(true)
    expect(store.firstTime('a')).toBe(false)
    expect(store.firstTime('b')).toBe(true)
  })
})
