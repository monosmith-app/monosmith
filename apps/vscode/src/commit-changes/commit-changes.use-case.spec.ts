import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { commitChanges } from './commit-changes.use-case.ts'

const PLAN: SmithRun = { code: 0, stdout: 'Plan: commit add app web\n  create apps/web/package.json\n', stderr: '' }

function fake (runs: SmithRun[], answer: boolean) {
  const calls: string[][] = []
  const asked: string[] = []

  return {
    calls,
    asked,
    deps: {
      run: async (args: string[]) => {
        calls.push(args)

        return runs.shift() ?? { code: 0, stdout: 'Committed 1a2b3c4d: chore: add app web (smith undo reverts it)\n', stderr: '' }
      },
      confirm: async (plan: string) => {
        asked.push(plan)

        return answer
      },
    },
  }
}

describe('commitChanges', () => {
  it('shows the files, asks, then commits with --yes under the name of the operation', async () => {
    const f = fake([PLAN], true)
    expect(await commitChanges(f.deps, 'add app web', 'operation')).toEqual({ kind: 'committed', lines: ['Committed 1a2b3c4d: chore: add app web (smith undo reverts it)'] })
    expect(f.calls).toEqual([['commit', 'add', 'app', 'web', '--only-operation', '--dry-run'], ['commit', 'add', 'app', 'web', '--only-operation', '--yes']])
    expect(f.asked[0]).toContain('apps/web/package.json')
  })

  it('passes --all only when the person chose everything', async () => {
    const f = fake([PLAN], true)
    await commitChanges(f.deps, 'add app web', 'all')
    expect(f.calls[0]).toEqual(['commit', 'add', 'app', 'web', '--all', '--dry-run'])
  })

  it('commits nothing when the person says no', async () => {
    const f = fake([PLAN], false)
    expect(await commitChanges(f.deps, 'add app web', 'operation')).toEqual({ kind: 'declined' })
    expect(f.calls).toHaveLength(1)
  })

  it('returns smith words when it refuses', async () => {
    const f = fake([{ code: 1, stdout: '', stderr: '✖ this workspace is not in a git work tree\n  → run git init, then commit again\n' }], true)
    expect(await commitChanges(f.deps, 'add app web', 'operation')).toMatchObject({ kind: 'refused' })
    expect(f.asked).toEqual([])
  })
})
