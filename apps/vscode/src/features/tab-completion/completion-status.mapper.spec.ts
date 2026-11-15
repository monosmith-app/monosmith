import { describe, expect, it } from 'vitest'
import { parseCompletionStatus } from './completion-status.mapper.ts'

describe('parseCompletionStatus', () => {
  it('reads what smith prints', () => {
    expect(parseCompletionStatus('{"shell":"zsh","wanted":true,"answer":""}\n')).toEqual({ shell: 'zsh', wanted: true, answer: '' })
  })

  it('says nothing for text that is not the status: an older smith, or an error', () => {
    expect(parseCompletionStatus('')).toBeUndefined()
    expect(parseCompletionStatus('unknown command "status"')).toBeUndefined()
    expect(parseCompletionStatus('{"shell":"zsh"}')).toBeUndefined()
    expect(parseCompletionStatus('null')).toBeUndefined()
  })
})
