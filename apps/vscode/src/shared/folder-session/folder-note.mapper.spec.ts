import { describe, expect, it } from 'vitest'
import { describeFolderNote } from './folder-note.mapper.ts'

describe('describeFolderNote', () => {
  it('has nothing to say of a ready folder', () => {
    expect(describeFolderNote({ phase: 'ready' })).toBe('')
  })

  it('names each state that is not ready', () => {
    expect(describeFolderNote({ phase: 'missing' })).toBe('smith not found')
    expect(describeFolderNote({ phase: 'untrusted' })).toBe('workspace not trusted')
    expect(describeFolderNote({ phase: 'starting' })).toContain('starting')
  })

  it('gives the first line of why smith failed', () => {
    expect(describeFolderNote({ phase: 'failed', error: 'boom\nstack' })).toBe('smith failed: boom')
  })
})
