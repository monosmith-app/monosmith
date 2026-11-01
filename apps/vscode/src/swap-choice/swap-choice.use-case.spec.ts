import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { swapArguments } from './swap-arguments.mapper.ts'
import type { SwapAnswer, SwapRequest } from './swap-choice.contract.ts'
import { swapChoice } from './swap-choice.use-case.ts'

const REQUEST: SwapRequest = { project: 'kit', axis: 'test', choice: 'vitest' }
const CHOICES: SmithRun = {
  code:   0,
  stdout: JSON.stringify([
    { axis: 'test', name: 'jest', description: 'Jest', language: 'node' },
    { axis: 'test', name: 'vitest', description: 'Vitest', language: 'node' },
  ]),
  stderr: '',
}
const PLAN: SmithRun = { code: 0, stdout: 'Swap test of kit: jest -> vitest\nDry run: nothing was changed.\n', stderr: '' }

function fake (answers: SmithRun[], answer: SwapAnswer, request: SwapRequest | null = REQUEST): { calls: string[][], applied: string[][], deps: Parameters<typeof swapChoice>[0] } {
  const calls: string[][] = []
  const applied: string[][] = []
  const queue = [...answers]

  return {
    calls,
    applied,
    deps: {
      run: async (args) => {
        calls.push(args)

        return queue.shift() ?? CHOICES
      },
      apply: async (args) => {
        applied.push(args)

        return { code: 0, stdout: 'Verified: test passed in 1s\nLeft to do by hand: a.test.ts uses jest.fn\n', stderr: '' }
      },
      parse:   (json) => JSON.parse(json) as never,
      ask:     async () => request ?? undefined,
      confirm: async () => answer,
    },
  }
}

describe('swapArguments', () => {
  it('plans with --dry-run and applies with --yes, adding --commit only on request', () => {
    expect(swapArguments(REQUEST, 'plan')).toEqual(['swap', 'kit', 'test', 'vitest', '--dry-run'])
    expect(swapArguments(REQUEST, 'apply')).toEqual(['swap', 'kit', 'test', 'vitest', '--yes'])
    expect(swapArguments(REQUEST, 'apply', true)).toEqual(['swap', 'kit', 'test', 'vitest', '--yes', '--commit'])
    expect(swapArguments(REQUEST, 'plan', true)).not.toContain('--commit')
  })
})

describe('swapChoice', () => {
  it('shows the plan first and swaps only on a yes', async () => {
    const f = fake([CHOICES, PLAN], 'swap')
    const outcome = await swapChoice(f.deps)
    expect(f.calls).toEqual([['choices', '--json'], ['swap', 'kit', 'test', 'vitest', '--dry-run']])
    expect(f.applied).toEqual([['swap', 'kit', 'test', 'vitest', '--yes']])
    expect(outcome).toMatchObject({ kind: 'swapped', committed: false })
  })

  it('commits when the person chose the commit button', async () => {
    const f = fake([CHOICES, PLAN], 'swap-commit')
    expect(await swapChoice(f.deps)).toMatchObject({ kind: 'swapped', committed: true })
    expect(f.applied[0]).toContain('--commit')
  })

  it('changes nothing when the person cancels', async () => {
    const f = fake([CHOICES, PLAN], 'cancel')
    expect(await swapChoice(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.applied).toEqual([])
    expect(await swapChoice(fake([CHOICES], 'swap', null).deps)).toEqual({ kind: 'cancelled' })
  })

  it('returns the words of smith when the plan is refused', async () => {
    const f = fake([CHOICES, { code: 1, stdout: '', stderr: '✖ the project already uses vitest on the test axis\n' }], 'swap')
    expect(await swapChoice(f.deps)).toEqual({ kind: 'refused', message: 'the project already uses vitest on the test axis' })
    expect(f.applied).toEqual([])
  })
})
