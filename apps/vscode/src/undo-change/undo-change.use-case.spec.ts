import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { undoLastChange } from './undo-change.use-case.ts'

const PLAN: SmithRun = { code: 0, stdout: 'Plan: undo add lib core (commit 6b0d49f4)\n  remove libs/core/go.mod\nDry run: nothing was changed.\n', stderr: '' }

function fake (runs: SmithRun[], answer: boolean) {
  const calls: string[][] = []
  const asked: string[] = []

  return {
    calls,
    asked,
    deps: {
      run: async (args: string[]) => {
        calls.push(args)

        return runs.shift() ?? { code: 0, stdout: 'reverted\n', stderr: '' }
      },
      confirm: async (plan: string) => {
        asked.push(plan)

        return answer
      },
    },
  }
}

describe('undoLastChange', () => {
  it('shows the dry run, asks, then reverts with --yes', async () => {
    const f = fake([PLAN], true)
    expect(await undoLastChange(f.deps)).toEqual({ kind: 'undone', lines: ['reverted'] })
    expect(f.calls).toEqual([['undo', '--dry-run'], ['undo', '--yes']])
    expect(f.asked[0]).toContain('Plan: undo add lib core')
  })

  it('reverts nothing when the person says no', async () => {
    const f = fake([PLAN], false)
    expect(await undoLastChange(f.deps)).toEqual({ kind: 'declined' })
    expect(f.calls).toHaveLength(1)
  })

  it('tells how to turn it on when smith has no operation to undo', async () => {
    const f = fake([{ code: 1, stdout: '', stderr: '✖ smith has no operation to undo here\n  → set commit_operations = true under [workspace] in smith.toml, then run a command such as smith add\n' }], true)
    expect(await undoLastChange(f.deps)).toEqual({
      kind:    'refused',
      message: 'smith has no operation to undo here Try: set commit_operations = true under [workspace] in smith.toml, then run a command such as smith add',
    })
    expect(f.asked).toEqual([])
  })

  it('returns the refusal of smith when the revert fails', async () => {
    const f = fake([PLAN, { code: 1, stdout: '', stderr: '✖ a later commit touches the same files\n' }], true)
    expect(await undoLastChange(f.deps)).toEqual({ kind: 'refused', message: 'a later commit touches the same files' })
  })
})
