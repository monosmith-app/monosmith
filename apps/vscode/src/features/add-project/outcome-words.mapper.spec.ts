import { describe, expect, it } from 'vitest'
import { outcomeWords } from './outcome-words.mapper.ts'

describe('outcomeWords', () => {
  it('says a written project was written, not refused or cancelled', () => {
    for (const kind of ['written', 'failing']) {
      expect(outcomeWords({ kind })).toContain('was written')
      expect(outcomeWords({ kind })).not.toMatch(/refused|cancelled/)
    }
  })

  it('keeps the words of smith for a failure', () => {
    expect(outcomeWords({ kind: 'refused', message: 'web exists Try: pick another name' })).toBe('stopped: web exists Try: pick another name')
  })

  it('says nothing was changed when the person cancelled or went to the composer', () => {
    expect(outcomeWords({ kind: 'cancelled' })).toContain('nothing was changed')
    expect(outcomeWords({ kind: 'composing' })).toContain('nothing was added')
  })

  it('falls back to the kind for a word it does not know', () => {
    expect(outcomeWords({ kind: 'klingon' })).toBe('klingon')
  })
})
