import { describe, expect, it } from 'vitest'
import { shouldOfferTabCompletion } from './tab-completion-offer.policy.ts'

describe('shouldOfferTabCompletion', () => {
  it('offers when smith says it is not set up and nothing was answered', () => {
    expect(shouldOfferTabCompletion({ wanted: true, remembered: undefined, askedThisSession: false })).toBe(true)
  })

  it('never offers what smith does not want: it is installed, or the person said no on the command line', () => {
    expect(shouldOfferTabCompletion({ wanted: false, remembered: undefined, askedThisSession: false })).toBe(false)
  })

  it('stays quiet after "Set up" and after "Don\'t ask again"', () => {
    expect(shouldOfferTabCompletion({ wanted: true, remembered: 'done', askedThisSession: false })).toBe(false)
    expect(shouldOfferTabCompletion({ wanted: true, remembered: 'never', askedThisSession: false })).toBe(false)
  })

  it('does not ask twice in one session after "Not now"', () => {
    expect(shouldOfferTabCompletion({ wanted: true, remembered: undefined, askedThisSession: true })).toBe(false)
  })
})
