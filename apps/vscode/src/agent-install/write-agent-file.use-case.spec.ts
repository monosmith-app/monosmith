import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { writeAgentFile } from './write-agent-file.use-case.ts'

const CURRENT: SmithRun = { code: 0, stdout: 'AGENTS.md is current\n', stderr: '' }
const MISSING: SmithRun = { code: 1, stdout: 'AGENTS.md is missing, run smith agent file\n', stderr: '✖ AGENTS.md is missing\n  → smith agent file\n' }
const WRITTEN: SmithRun = { code: 0, stdout: 'wrote AGENTS.md\nwrote CLAUDE.md\n', stderr: '' }

function fake (runs: SmithRun[], answer: boolean) {
  const calls: string[][] = []
  const asked: string[][] = []

  return {
    calls,
    asked,
    deps: {
      run: async (args: string[]) => {
        calls.push(args)

        return runs.shift() ?? WRITTEN
      },
      confirm: async (lines: string[]) => {
        asked.push(lines)

        return answer
      },
    },
  }
}

describe('writeAgentFile', () => {
  it('asks nothing and writes nothing when the files are current', async () => {
    const f = fake([CURRENT], true)
    expect(await writeAgentFile(f.deps)).toEqual({ kind: 'current' })
    expect(f.calls).toEqual([['agent', 'file', '--check']])
    expect(f.asked).toEqual([])
  })

  it('names what is out of date, asks, then writes', async () => {
    const f = fake([MISSING, WRITTEN], true)
    expect(await writeAgentFile(f.deps)).toEqual({ kind: 'written', lines: ['wrote AGENTS.md', 'wrote CLAUDE.md'] })
    expect(f.asked[0]).toEqual(['AGENTS.md is missing, run smith agent file'])
    expect(f.calls[1]).toEqual(['agent', 'file'])
  })

  it('writes nothing when the person says no', async () => {
    const f = fake([MISSING], false)
    expect(await writeAgentFile(f.deps)).toEqual({ kind: 'declined' })
    expect(f.calls).toHaveLength(1)
  })

  it('skips the question when the person already agreed', async () => {
    const f = fake([MISSING, WRITTEN], false)
    const outcome = await writeAgentFile(f.deps, true)
    expect(outcome.kind).toBe('written')
    expect(f.asked).toEqual([])
  })

  it('returns the words of smith when the write fails', async () => {
    const f = fake([MISSING, { code: 1, stdout: '', stderr: '✖ AGENTS.md is read-only\n  → make it writable\n' }], true)
    expect(await writeAgentFile(f.deps)).toEqual({ kind: 'refused', message: 'AGENTS.md is read-only Try: make it writable' })
  })

  it('refuses when smith cannot check at all', async () => {
    const f = fake([{ code: 1, stdout: '', stderr: '✖ smith.toml not found\n  → run smith in a workspace\n' }], true)
    expect(await writeAgentFile(f.deps)).toEqual({ kind: 'refused', message: 'smith.toml not found Try: run smith in a workspace' })
  })
})
