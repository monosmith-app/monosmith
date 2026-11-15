import { describe, expect, it } from 'vitest'
import { lastLines } from './stderr-tail.algorithm.ts'

describe('lastLines', () => {
  it('keeps the last lines only', () => {
    expect(lastLines('a\nb\nc\nd', 2)).toBe('c\nd')
  })

  it('skips empty lines and handles CRLF', () => {
    expect(lastLines('a\r\n\r\nb\r\n', 5)).toBe('a\nb')
  })

  it('gives an empty text for no output', () => {
    expect(lastLines('', 3)).toBe('')
  })
})
