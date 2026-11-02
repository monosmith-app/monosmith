import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { forcedConsequences, parseUnchangedPackages } from './forced-release.mapper.ts'
import { releaseArguments } from './release-arguments.mapper.ts'
import { releaseWorkspace } from './release-workspace.use-case.ts'

const ok = (stdout: string): SmithRun => ({ code: 0, stdout, stderr: '' })
const NOTHING = '{"dryRun":true,"results":[],"unchanged":["mvd","core"]}'

function fake (answers: { forced?: string[], confirm?: boolean }) {
  const calls: string[][] = []
  const applied: string[][] = []
  const confirmed: string[][] = []

  return {
    calls,
    applied,
    confirmed,
    deps: {
      run: async (args: string[]): Promise<SmithRun> => {
        calls.push(args)

        return ok(args.includes('--json') ? NOTHING : 'plan\n')
      },
      apply: async (args: string[]): Promise<SmithRun> => {
        applied.push(args)

        return ok('Released mvd 0.14.1\n')
      },
      chooseHold:   async () => [],
      chooseBump:   async () => [],
      chooseForced: async (_unchanged: string[], _nothing: boolean) => answers.forced,
      confirm:      async (_plan: string, choices: { forced?: string[] }) => {
        confirmed.push(choices.forced ?? [])

        return answers.confirm ?? true
      },
    },
  }
}

describe('parseUnchangedPackages', () => {
  it('reads the names, and none when smith lists none', () => {
    expect(parseUnchangedPackages(NOTHING)).toEqual(['mvd', 'core'])
    expect(parseUnchangedPackages('{"results":[]}')).toEqual([])
  })
})

describe('forcedConsequences', () => {
  it('names the package, the tag at HEAD and the publish; says nothing without one', () => {
    expect(forcedConsequences(['mvd=patch'])).toContain('mvd: no commit releases it. The new tag points at HEAD')
    expect(forcedConsequences([])).toBe('')
  })
})

describe('releaseArguments with a forced package', () => {
  it('adds the bump and --force', () => {
    expect(releaseArguments({ hold: [], bump: [], forced: ['mvd=patch'] }, 'publish')).toEqual(['release', '--yes', '--bump', 'mvd=patch', '--force'])
    expect(releaseArguments({ hold: [], bump: [], forced: ['mvd=patch'] }, 'plan')).toEqual(['release', '--dry-run', '--bump', 'mvd=patch', '--force'])
  })
})

describe('releaseWorkspace without a releasing commit', () => {
  it('offers the path when the plan is empty, plans with --force, asks, then publishes', async () => {
    const f = fake({ forced: ['mvd=patch'] })
    expect(await releaseWorkspace(f.deps)).toMatchObject({ kind: 'released' })
    expect(f.calls.at(-1)).toEqual(['release', '--dry-run', '--bump', 'mvd=patch', '--force'])
    expect(f.confirmed).toEqual([['mvd=patch']])
    expect(f.applied).toEqual([['release', '--yes', '--bump', 'mvd=patch', '--force']])
  })

  it('says there is nothing to release when the person picks no package', async () => {
    const f = fake({ forced: [] })
    expect(await releaseWorkspace(f.deps)).toMatchObject({ kind: 'nothing' })
    expect(f.applied).toEqual([])
  })

  it('publishes nothing when the person cancels or says no', async () => {
    expect(await releaseWorkspace(fake({ forced: undefined }).deps)).toEqual({ kind: 'cancelled' })
    const f = fake({ forced: ['mvd=patch'], confirm: false })
    expect(await releaseWorkspace(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.applied).toEqual([])
  })
})
