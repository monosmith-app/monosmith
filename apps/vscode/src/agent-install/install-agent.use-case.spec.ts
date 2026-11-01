import { describe, expect, it } from 'vitest'
import { installAgent } from './install-agent.use-case.ts'
import type { SmithRun } from '../smith-run/index.ts'
import type { InstallPreview } from './install-run.contract.ts'

const PREVIEW_RUN: SmithRun = { code: 1, stdout: '● Add .vscode/mcp.json\n--- a\n+++ b\n', stderr: '✖ nothing was written: there is no terminal to ask on\n' }
const WRITTEN_RUN: SmithRun = { code: 0, stdout: '✓ Wrote .vscode/mcp.json\n', stderr: '' }

function deps (runs: SmithRun[], answer: boolean) {
  const calls: string[][] = []
  const asked: InstallPreview[] = []

  return {
    calls,
    asked,
    deps: {
      run: async (args: string[]) => {
        calls.push(args)

        return runs.shift() ?? WRITTEN_RUN
      },
      confirm: async (preview: InstallPreview) => {
        asked.push(preview)

        return answer
      },
    },
  }
}

describe('installAgent', () => {
  it('previews without --yes, asks, then writes with --yes', async () => {
    const fake = deps([PREVIEW_RUN, WRITTEN_RUN], true)
    expect(await installAgent(fake.deps, 'agent')).toEqual({ kind: 'written', lines: ['Wrote .vscode/mcp.json'] })
    expect(fake.calls).toEqual([['agent', 'install', '--claude', '--vscode'], ['agent', 'install', '--claude', '--vscode', '--yes']])
    expect(fake.asked[0]?.actions).toEqual(['Add .vscode/mcp.json'])
  })

  it('writes nothing when the person says no', async () => {
    const fake = deps([PREVIEW_RUN], false)
    expect(await installAgent(fake.deps, 'mcp')).toEqual({ kind: 'declined' })
    expect(fake.calls).toEqual([['mcp', 'install', '--vscode']])
  })

  it('does not ask again when the person already said yes', async () => {
    const fake = deps([PREVIEW_RUN, WRITTEN_RUN], false)
    const outcome = await installAgent(fake.deps, 'mcp', true)
    expect(outcome.kind).toBe('written')
    expect(fake.asked).toEqual([])
  })

  it('does not ask or write when smith refuses the preview', async () => {
    const fake = deps([{ code: 1, stdout: '', stderr: '✖ .mcp.json was changed after smith wrote it\n' }], true)
    expect(await installAgent(fake.deps, 'agent')).toEqual({ kind: 'refused', message: '.mcp.json was changed after smith wrote it' })
    expect(fake.calls).toHaveLength(1)
    expect(fake.asked).toEqual([])
  })

  it('does nothing when the files are current', async () => {
    const fake = deps([{ code: 0, stdout: '✓ The MonoSmith agent is already installed (3 files)\n', stderr: '' }], true)
    expect(await installAgent(fake.deps, 'agent')).toEqual({ kind: 'current' })
    expect(fake.calls).toHaveLength(1)
  })
})
