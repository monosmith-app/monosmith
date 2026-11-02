import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { parsePlannedPackages, releaseArguments } from './release-arguments.mapper.ts'
import type { PlannedPackage } from './release-flow.contract.ts'
import { releaseWorkspace } from './release-workspace.use-case.ts'

/** Real output of `smith release --dry-run --json` (2.0.0-beta.23) for a workspace with one new package. */
const planJson = readFileSync(new URL('fixtures/release-plan.fixture.json', import.meta.url), 'utf8')
const ok = (stdout: string): SmithRun => ({ code: 0, stdout, stderr: '' })
const PLAN_TEXT = 'Dry run: nothing is published or tagged.\nsdk unreleased -> 0.1.0\n'

function fake (answers: { hold?: string[], bump?: string[], confirm?: boolean }, json = planJson) {
  const calls: string[][] = []
  const applied: string[][] = []
  const events: string[] = []

  return {
    calls,
    applied,
    events,
    deps: {
      run: async (args: string[]): Promise<SmithRun> => {
        calls.push(args)

        return ok(args.includes('--json') ? json : PLAN_TEXT)
      },
      apply: async (args: string[]): Promise<SmithRun> => {
        applied.push(args)

        return ok('Released sdk 0.1.0\n')
      },
      chooseHold: async (_plan: string, _packages: PlannedPackage[]) => {
        events.push('hold')

        return answers.hold
      },
      chooseBump: async (packages: PlannedPackage[]) => {
        events.push(`bump:${packages.map((item) => item.name).join(',')}`)

        return answers.bump
      },
      confirm: async () => {
        events.push('confirm')

        return answers.confirm ?? true
      },
    },
  }
}

describe('parsePlannedPackages', () => {
  it('reads the packages of the real plan of smith', () => {
    expect(parsePlannedPackages(planJson)).toEqual([{ name: 'sdk', next: '0.1.0', level: 'minor' }])
  })

  it('reads no package from an empty plan', () => {
    expect(parsePlannedPackages('{"dryRun":true,"results":[]}')).toEqual([])
  })
})

describe('releaseArguments', () => {
  it('joins the held packages and repeats --bump', () => {
    expect(releaseArguments({ hold: ['a', 'b'], bump: ['a=major', 'c=patch'] }, 'publish')).toEqual(['release', '--yes', '--hold', 'a,b', '--bump', 'a=major', '--bump', 'c=patch'])
    expect(releaseArguments({ hold: [], bump: [] }, 'plan')).toEqual(['release', '--dry-run'])
  })
})

describe('releaseWorkspace', () => {
  it('plans, asks the hold, the bump and the final question, then publishes with --yes', async () => {
    const f = fake({ hold: [], bump: [] })
    expect(await releaseWorkspace(f.deps)).toEqual({ kind: 'released', lines: ['Released sdk 0.1.0'] })
    expect(f.events).toEqual(['hold', 'bump:sdk', 'confirm'])
    expect(f.applied).toEqual([['release', '--yes']])
  })

  it('passes the held packages and the bumps to the plan and to the release', async () => {
    const f = fake({ hold: ['sdk'], bump: ['api=major'] })
    await releaseWorkspace(f.deps)
    expect(f.events[1]).toBe('bump:')
    expect(f.calls.at(-1)).toEqual(['release', '--dry-run', '--hold', 'sdk', '--bump', 'api=major'])
    expect(f.applied).toEqual([['release', '--yes', '--hold', 'sdk', '--bump', 'api=major']])
  })

  it('publishes nothing when the person says no to the final question', async () => {
    const f = fake({ hold: [], bump: [], confirm: false })
    expect(await releaseWorkspace(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.applied).toEqual([])
  })

  it('stops at the first cancelled question', async () => {
    const f = fake({ hold: undefined })
    expect(await releaseWorkspace(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.events).toEqual(['hold'])
  })

  it('says why when there is nothing to release, and asks nothing', async () => {
    const f = fake({}, '{"dryRun":true,"results":[]}')
    expect(await releaseWorkspace(f.deps)).toMatchObject({ kind: 'nothing' })
    expect(f.events).toEqual([])
  })
})
