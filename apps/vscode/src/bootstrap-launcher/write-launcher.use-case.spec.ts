import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { writeLauncher } from './write-launcher.use-case.ts'

function fake (run: SmithRun, answer: boolean) {
  const calls: string[][] = []
  const shown: string[] = []

  return {
    calls,
    shown,
    deps: {
      run: async (args: string[]) => {
        calls.push(args)

        return run
      },
      confirm: async (plan: string) => {
        shown.push(plan)

        return answer
      },
    },
  }
}

describe('writeLauncher', () => {
  it('writes the launcher after a yes', async () => {
    const f = fake({ code: 0, stdout: 'wrote tools/smith/smith\nwrote tools/smith/smith.cmd\n', stderr: '' }, true)
    expect(await writeLauncher(f.deps)).toEqual({ kind: 'written', lines: ['wrote tools/smith/smith', 'wrote tools/smith/smith.cmd'] })
    expect(f.calls).toEqual([['bootstrap', '--dry-run'], ['bootstrap', '--yes']])
  })

  it('shows the plan smith printed before it asks', async () => {
    const f = fake({ code: 0, stdout: 'Plan: write the smith launcher\n  change tools/smith/smith\n', stderr: '' }, true)
    await writeLauncher(f.deps)
    expect(f.shown).toEqual(['Plan: write the smith launcher\n  change tools/smith/smith'])
  })

  it('writes nothing after a no', async () => {
    const f = fake({ code: 0, stdout: '', stderr: '' }, false)
    expect(await writeLauncher(f.deps)).toEqual({ kind: 'declined' })
    expect(f.calls).toEqual([['bootstrap', '--dry-run']])
  })

  it('returns the refusal of smith', async () => {
    const f = fake({ code: 1, stdout: '', stderr: '✖ not a smith workspace\n' }, true)
    expect(await writeLauncher(f.deps)).toEqual({ kind: 'refused', message: 'not a smith workspace' })
  })
})
