import { describe, expect, it } from 'vitest'
import { fixFinding, type FixFindingDeps } from './fix-finding.use-case.ts'
import type { DoctorFinding } from './doctor-report.contract.ts'

const defaults: DoctorFinding = {
  check: 'the workspace uses the latest defaults',
  ok:    false,
  fix:   { description: 'Move the workspace to the latest defaults' },
}

const derive: DoctorFinding = {
  check: 'derived file a.json matches',
  ok:    false,
  fix:   { description: 'Settle the derived file a.json', choices: [{ question: 'Who owns a.json?', options: [{ label: 'smith' }, { label: 'You' }] }] },
}

const unanswered = async (): Promise<{ code: number, stdout: string, stderr: string }> => ({ code: 1, stdout: '● review\n\nnot applied: x needs your decision. Run one of:\n  smith doctor --fix --only x --yes\n', stderr: '' })

function deps (over: Partial<FixFindingDeps> = {}, ran: string[][] = []): FixFindingDeps {
  return {
    run: async (args) => {
      ran.push(args)

      return { code: 0, stdout: '● review\n  step 1: $ smith upgrade --defaults\n', stderr: '' }
    },
    pick:    async () => 0,
    confirm: async () => true,
    ...over,
  }
}

describe('fixFinding', () => {
  it('shows the review first and applies with --yes only after a yes', async () => {
    const ran: string[][] = []
    const shown: string[] = []
    const confirm = async (review: string): Promise<boolean> => {
      shown.push(review)

      return true
    }
    const outcome = await fixFinding(deps({ confirm }, ran), defaults)

    expect(outcome.kind).toBe('fixed')
    expect(shown[0]).toContain('smith upgrade --defaults')
    expect(ran).toEqual([
      ['doctor', '--fix', '--only', defaults.fix?.description],
      ['doctor', '--fix', '--only', defaults.fix?.description, '--yes'],
    ])
  })

  it('does not show the terminal command smith prints for a fix nobody could answer', async () => {
    const shown: string[] = []
    const confirm = async (review: string): Promise<boolean> => {
      shown.push(review)

      return false
    }
    await fixFinding(deps({ run: unanswered, confirm }), defaults)

    expect(shown[0]).toBe('● review')
  })

  it('changes nothing when the person declines', async () => {
    const ran: string[][] = []
    const outcome = await fixFinding(deps({ confirm: async () => false }, ran), defaults)

    expect(outcome).toEqual({ kind: 'declined' })
    expect(ran.flat()).not.toContain('--yes')
  })

  it('asks every question and passes the answers by number', async () => {
    const ran: string[][] = []
    await fixFinding(deps({ pick: async () => 1 }, ran), derive)

    expect(ran[1]).toEqual(['doctor', '--fix', '--only', derive.fix?.description, '--choose', '2', '--yes'])
  })

  it('stops quietly when the person cancels a question', async () => {
    const ran: string[][] = []

    expect(await fixFinding(deps({ pick: async () => undefined }, ran), derive)).toEqual({ kind: 'declined' })
    expect(ran).toEqual([])
  })

  it('reports an applied fix while other checks still fail', async () => {
    let calls = 0
    const outcome = await fixFinding(deps({
      run: async () => ({ code: ++calls === 1 ? 0 : 1, stdout: calls === 1 ? '● review' : 'after fixing: 3 passed, 1 failed, 0 warning(s)', stderr: '' }),
    }), defaults)

    expect(outcome).toMatchObject({ kind: 'fixed', incomplete: true })
  })

  it('says why when smith refuses', async () => {
    const outcome = await fixFinding(deps({ run: async () => ({ code: 1, stdout: '', stderr: '× not here\n  → do this' }) }), defaults)

    expect(outcome).toEqual({ kind: 'refused', message: 'not here Try: do this' })
  })

  it('has nothing to do for a check without a fix or one that passed', async () => {
    expect(await fixFinding(deps(), { check: 'x', ok: false })).toEqual({ kind: 'nothing' })
    expect(await fixFinding(deps(), { ...defaults, ok: true })).toEqual({ kind: 'nothing' })
  })
})
