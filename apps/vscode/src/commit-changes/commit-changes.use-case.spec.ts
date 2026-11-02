import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { commitChanges, readCommitPlan, type CommitPlan, type CommitScope } from './commit-changes.use-case.ts'

const PLAN: SmithRun = { code: 0, stdout: 'Plan: commit add app web\n  Goes into one commit (chore: add app web)\n    create  apps/web/package.json\nDry run: nothing was changed.\n', stderr: '' }
const WITH_EXTRAS: SmithRun = {
  code:   0,
  stdout: 'Plan: commit add app web\n    create  apps/web/package.json\nThe last operation changed 1 file(s). Git shows 2 more that it did not change:\n  mine.txt\n  notes/todo.md\n--only-operation commits the first set, --all both. Dry run: nothing was changed.\n',
  stderr: '',
}

function fake (runs: SmithRun[], answer: CommitScope | undefined) {
  const calls: string[][] = []
  const asked: CommitPlan[] = []

  return {
    calls,
    asked,
    deps: {
      run: async (args: string[]) => {
        calls.push(args)

        return runs.shift() ?? { code: 0, stdout: 'Committed 1a2b3c4d: chore: add app web (smith undo reverts it)\n', stderr: '' }
      },
      choose: async (plan: CommitPlan) => {
        asked.push(plan)

        return answer
      },
    },
  }
}

describe('readCommitPlan', () => {
  it('lists the files git shows that the operation did not change', () => {
    expect(readCommitPlan(WITH_EXTRAS.stdout).extras).toEqual(['mine.txt', 'notes/todo.md'])
  })

  it('has no extras when the operation changed everything', () => {
    expect(readCommitPlan(PLAN.stdout).extras).toEqual([])
  })
})

describe('commitChanges', () => {
  it('shows the plan first, then commits only the files of the operation with --yes', async () => {
    const f = fake([PLAN], 'operation')
    expect(await commitChanges(f.deps, 'add app web')).toEqual({ kind: 'committed', lines: ['Committed 1a2b3c4d: chore: add app web (smith undo reverts it)'] })
    expect(f.calls).toEqual([['commit', 'add', 'app', 'web', '--dry-run'], ['commit', 'add', 'app', 'web', '--only-operation', '--yes']])
    expect(f.asked[0]?.text).toContain('apps/web/package.json')
  })

  it('works for the last operation when no words are given, and passes the extras to the question', async () => {
    const f = fake([WITH_EXTRAS], 'all')
    await commitChanges(f.deps)
    expect(f.calls).toEqual([['commit', '--dry-run'], ['commit', '--all', '--yes']])
    expect(f.asked[0]?.extras).toEqual(['mine.txt', 'notes/todo.md'])
  })

  it('commits nothing when the person cancels', async () => {
    const f = fake([PLAN], undefined)
    expect(await commitChanges(f.deps, 'add app web')).toEqual({ kind: 'declined' })
    expect(f.calls).toHaveLength(1)
  })

  it('returns smith words when it refuses', async () => {
    const f = fake([{ code: 1, stdout: '', stderr: '✖ this workspace is not in a git work tree\n  → run git init, then commit again\n' }], 'operation')
    expect(await commitChanges(f.deps, 'add app web')).toMatchObject({ kind: 'refused' })
    expect(f.asked).toEqual([])
  })
})
