import { describe, expect, it } from 'vitest'
import { dependencyQuestion, lintQuestion, toSplitAnswer } from './split-choices.policy.ts'

describe('dependencyQuestion', () => {
  it('names the package and offers the project first, because it changes the least', () => {
    const question = dependencyQuestion('supertest')

    expect(question.title).toContain('supertest')
    expect(question.choices.map((c) => c.value)).toEqual(['project', 'workspace'])
    expect(question.choices.every((c) => c.label !== '' && c.description !== '')).toBe(true)
  })
})

describe('lintQuestion', () => {
  it('offers keeping the own setup first and the shared config second', () => {
    const question = lintQuestion()

    expect(question.choices.map((c) => c.value)).toEqual(['keep', 'shared'])
    expect(question.choices[1]?.description).toContain('.pre-smith')
  })
})

describe('toSplitAnswer', () => {
  it('sends only the dependencies the person put in the workspace', () => {
    expect(toSplitAnswer([{ name: 'nock', goesTo: 'project' }, { name: 'supertest', goesTo: 'workspace' }], 'shared')).toEqual({ hoist: ['supertest'], eslint: 'shared' })
  })

  it('has no ESLint answer when the question was not asked', () => {
    expect(toSplitAnswer([], undefined)).toEqual({ hoist: [], eslint: undefined })
  })
})
