import { describe, expect, it } from 'vitest'
import { optOutOfAgent } from './opt-out-agent.use-case.ts'

describe('optOutOfAgent', () => {
  it('asks smith config to record the opt-out', async () => {
    const calls: string[][] = []
    const failure = await optOutOfAgent(async (args) => {
      calls.push(args)

      return { code: 0, stdout: '✔ agent.enabled = false\n', stderr: '' }
    })
    expect(failure).toBeUndefined()
    expect(calls).toEqual([['config', 'set', 'agent.enabled', 'false']])
  })

  it('returns what smith said when it could not', async () => {
    expect(await optOutOfAgent(async () => ({ code: 1, stdout: '', stderr: '✖ smith.toml is read-only\n' }))).toBe('smith.toml is read-only')
  })
})
