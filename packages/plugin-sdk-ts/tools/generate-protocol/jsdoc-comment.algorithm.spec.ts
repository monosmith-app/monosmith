import { describe, expect, it } from 'vitest'
import { jsdocComment } from './jsdoc-comment.algorithm.ts'

describe('jsdocComment', () => {
  it('writes nothing for no text and one line for a short text', () => {
    expect(jsdocComment(undefined)).toBe('')
    expect(jsdocComment('  ')).toBe('')
    expect(jsdocComment('Short.', '  ')).toBe('  /** Short. */\n')
  })

  it('wraps a long text, keeps paragraphs and escapes a terminator', () => {
    const words = Array.from({ length: 40 }, (_, i) => `word${i}`).join(' ')
    const block = jsdocComment(`${words}\n\nSecond */ paragraph.`)
    const lines = block.split('\n')
    expect(lines[0]).toBe('/**')
    expect(lines.at(-2)).toBe(' */')
    expect(lines.every(line => line.length <= 100)).toBe(true)
    expect(block).toContain(' *\n * Second *\\/ paragraph.\n')
  })
})
