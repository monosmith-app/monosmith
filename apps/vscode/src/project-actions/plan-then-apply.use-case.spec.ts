import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { kindArguments, kindPlanArguments, otherKinds, removeArguments, removePlanArguments, unlinkArguments } from './project-action-arguments.mapper.ts'
import { planThenApply } from './plan-then-apply.use-case.ts'

function fake<Answer> (planRun: SmithRun, answer: Answer | undefined) {
  const calls: string[][] = []
  const applied: string[][] = []
  const asked: string[] = []

  return {
    calls,
    applied,
    asked,
    deps: {
      run: async (args: string[]) => {
        calls.push(args)

        return planRun
      },
      apply: async (args: string[]) => {
        applied.push(args)

        return { code: 0, stdout: 'Done: removed web\n', stderr: '' }
      },
      confirm: async (plan: string) => {
        asked.push(plan)

        return answer
      },
    },
  }
}

describe('planThenApply', () => {
  const plan: SmithRun = { code: 0, stdout: 'Plan: remove web\n  delete apps/web\n', stderr: '' }

  it('shows the plan of the dry run, then applies the answer with --yes', async () => {
    const f = fake(plan, 'keep-copy' as const)
    const outcome = await planThenApply(f.deps, removePlanArguments('web'), (answer) => removeArguments('web', answer))
    expect(outcome).toEqual({ kind: 'done', lines: ['Done: removed web'] })
    expect(f.calls).toEqual([['remove', 'web', '--dry-run']])
    expect(f.asked[0]).toContain('delete apps/web')
    expect(f.applied).toEqual([['remove', 'web', '--keep-copy', '--yes']])
  })

  it('passes --delete only when the person chose to delete', async () => {
    const f = fake(plan, 'delete' as const)
    await planThenApply(f.deps, removePlanArguments('web'), (answer) => removeArguments('web', answer))
    expect(f.applied).toEqual([['remove', 'web', '--delete', '--yes']])
  })

  it('changes nothing when the person cancels', async () => {
    const f = fake(plan, undefined)
    expect(await planThenApply(f.deps, kindPlanArguments('core', 'package'), () => kindArguments('core', 'package'))).toEqual({ kind: 'cancelled' })
    expect(f.applied).toEqual([])
  })

  it('refuses with smith words when the plan fails, without asking', async () => {
    const f = fake({ code: 1, stdout: '', stderr: '✖ core is imported by api\n  → pass --force\n' }, true)
    expect(await planThenApply(f.deps, kindPlanArguments('core', 'app'), () => kindArguments('core', 'app'))).toMatchObject({ kind: 'refused' })
    expect(f.asked).toEqual([])
  })

  it('asks with no plan when smith has no plan command (unlink)', async () => {
    const f = fake(plan, true)
    await planThenApply(f.deps, undefined, () => unlinkArguments('web', 'core'))
    expect(f.calls).toEqual([])
    expect(f.applied).toEqual([['uninstall', '-w', 'web', 'workspace:core', '--yes']])
  })
})

describe('otherKinds', () => {
  it('lists the kinds a project is not', () => {
    expect(otherKinds('app')).toEqual(['lib', 'package'])
    expect(otherKinds('package')).toEqual(['app', 'lib'])
  })
})
