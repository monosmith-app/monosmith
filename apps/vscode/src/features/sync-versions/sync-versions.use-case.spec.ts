import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { syncVersions, type SyncDeps, type SyncRule } from './sync-versions.use-case.ts'

function setup (rule: SyncRule | undefined, answer: boolean, fail = false): { deps: SyncDeps, calls: string[][], shown: string[] } {
  const calls: string[][] = []
  const shown: string[] = []

  return {
    calls,
    shown,
    deps: {
      run: async (args): Promise<SmithRun> => {
        calls.push(args)

        return fail ? { code: 1, stdout: '', stderr: 'smith: the tree is not clean' } : { code: 0, stdout: 'axios will be ^1.9.0\n', stderr: '' }
      },
      choose:  async () => rule,
      confirm: async (plan) => {
        shown.push(plan)

        return answer
      },
    },
  }
}

describe('syncVersions', () => {
  it('shows the plan for the chosen rule first and writes it with --yes only after a yes', async () => {
    const f = setup('highest', true)
    const outcome = await syncVersions(f.deps)
    expect(outcome.kind).toBe('synced')
    expect(f.calls).toEqual([['sync', '--prefer', 'highest', '--dry-run'], ['sync', '--prefer', 'highest', '--yes']])
    expect(f.shown).toEqual(['axios will be ^1.9.0'])
  })

  it('writes nothing when the person does not choose or says no', async () => {
    const none = setup(undefined, true)
    expect(await syncVersions(none.deps)).toEqual({ kind: 'declined' })
    expect(none.calls).toEqual([])
    const no = setup('lowest', false)
    expect(await syncVersions(no.deps)).toEqual({ kind: 'declined' })
    expect(no.calls).toEqual([['sync', '--prefer', 'lowest', '--dry-run']])
  })

  it('returns smith\'s words when the plan is refused', async () => {
    const f = setup('installed', true, true)
    const outcome = await syncVersions(f.deps)
    expect(outcome).toMatchObject({ kind: 'refused' })
    expect(f.calls).toHaveLength(1)
  })
})
