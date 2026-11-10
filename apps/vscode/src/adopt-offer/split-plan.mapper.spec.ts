import { describe, expect, it } from 'vitest'
import { parseSplitPlan } from './split-plan.mapper.ts'

describe('parseSplitPlan', () => {
  it('reads the dependencies smith cannot place and the ESLint question from the move plan', () => {
    const report = JSON.stringify({ blockers: [], movePlan: { dir: 'apps/web', unsure: ['nock', 'supertest'], lint: true, hoist: ['jest'] } })

    expect(parseSplitPlan(report)).toEqual({ unsure: ['nock', 'supertest'], lint: true })
  })

  it('asks nothing when the plan has no question', () => {
    expect(parseSplitPlan(JSON.stringify({ movePlan: { dir: 'apps/web' } }))).toEqual({ unsure: [], lint: false })
  })

  it('ignores a value that is not a name', () => {
    expect(parseSplitPlan(JSON.stringify({ movePlan: { unsure: ['nock', 3, null], lint: 'yes' } }))).toEqual({ unsure: ['nock'], lint: false })
  })

  it('has no plan when smith printed none, or no JSON', () => {
    expect(parseSplitPlan(JSON.stringify({ blockers: [] }))).toBeUndefined()
    expect(parseSplitPlan('smith adopt: something went wrong')).toBeUndefined()
    expect(parseSplitPlan('null')).toBeUndefined()
  })
})
