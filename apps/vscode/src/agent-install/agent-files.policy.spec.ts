import { describe, expect, it } from 'vitest'
import { AGENT_FILES_CHECK, judgeAgentFiles, shouldOfferAgent } from './agent-files.policy.ts'

describe('judgeAgentFiles', () => {
  it('says missing when the check warns', () => {
    expect(judgeAgentFiles({ findings: [{ check: AGENT_FILES_CHECK, ok: false, warning: true, detail: '.mcp.json' }] })).toBe('missing')
  })

  it('says current when the check passes', () => {
    expect(judgeAgentFiles({ findings: [{ check: AGENT_FILES_CHECK, ok: true, warning: true }] })).toBe('current')
  })

  it('says unknown when the check is absent: an older smith, or the opt-out', () => {
    expect(judgeAgentFiles({ findings: [{ check: 'the cache directory is ignored by git', ok: false }] })).toBe('unknown')
    expect(judgeAgentFiles({})).toBe('unknown')
  })
})

describe('shouldOfferAgent', () => {
  it('offers once, when the files are missing', () => {
    expect(shouldOfferAgent({ state: 'missing', alreadyAsked: false })).toBe(true)
    expect(shouldOfferAgent({ state: 'missing', alreadyAsked: true })).toBe(false)
  })

  it('never offers what is there or cannot be told', () => {
    expect(shouldOfferAgent({ state: 'current', alreadyAsked: false })).toBe(false)
    expect(shouldOfferAgent({ state: 'unknown', alreadyAsked: false })).toBe(false)
  })
})
