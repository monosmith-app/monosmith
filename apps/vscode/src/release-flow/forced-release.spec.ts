import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { firstVersionCommand, forcedConsequences, parseNeverReleasedPackages, parseUnchangedPackages } from './forced-release.mapper.ts'
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

  it('reads the packages never released, and the commit that gives one its first version', () => {
    expect(parseNeverReleasedPackages('{"unchanged":["a"],"neverReleased":["a"]}')).toEqual(['a'])
    expect(parseNeverReleasedPackages(NOTHING)).toEqual([])
    expect(firstVersionCommand('a')).toBe('git commit --allow-empty -m "version(a)[1.0.0]"')
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
    // The packages no commit releases go with the outcome, so the notice can offer the forced path.
    expect(await releaseWorkspace(f.deps)).toMatchObject({ kind: 'nothing', unchanged: ['mvd', 'core'] })
    expect(f.applied).toEqual([])
  })

  it('does not offer a package that was never released, and tells the picker which ones they are', async () => {
    const f = fake({ forced: [] })
    const asked: Array<{ unchanged: string[], never: string[] }> = []
    f.deps.chooseForced = async (unchanged: string[], _nothing: boolean, never: string[] = []) => {
      asked.push({ unchanged, never })

      return []
    }
    f.deps.run = async (args: string[]) => ok(args.includes('--json') ? '{"dryRun":true,"results":[],"unchanged":["mvd","fresh"],"neverReleased":["fresh"]}' : 'plan\n')
    await releaseWorkspace(f.deps)
    expect(asked).toEqual([{ unchanged: ['mvd'], never: ['fresh'] }])
  })

  it('publishes nothing when the person cancels or says no', async () => {
    expect(await releaseWorkspace(fake({ forced: undefined }).deps)).toEqual({ kind: 'cancelled' })
    const f = fake({ forced: ['mvd=patch'], confirm: false })
    expect(await releaseWorkspace(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.applied).toEqual([])
  })
})
