import { describe, expect, it } from 'vitest'
import { judgeProtocol } from './protocol-range.policy.ts'

describe('judgeProtocol', () => {
  it('accepts the protocol the extension speaks', () => {
    expect(judgeProtocol(1)).toBe('supported')
  })

  it('refuses an older one', () => {
    expect(judgeProtocol(0)).toBe('older')
  })

  it('works with a newer one but says so', () => {
    expect(judgeProtocol(2)).toBe('newer')
  })
})
