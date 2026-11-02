import { describe, expect, it } from 'vitest'
import { appliesWhen, askableQuestions, parseNewQuestions } from './new-questions.mapper.ts'

const JSON_TEXT = JSON.stringify([
  { key: 'ci', prompt: 'Which CI?', default: 'github', flag: 'ci', many: true, options: [{ value: 'github', label: 'GitHub Actions', hint: 'default' }, { value: 'azure', label: 'Azure Pipelines' }] },
  { key: 'node-stack', prompt: 'Which stack?', default: 'node-jest', flag: 'default-stack', when: { language: 'node' }, options: [{ value: 'node-jest', label: 'Jest' }] },
  { key: 'more', prompt: 'More?', default: 'no', options: [{ value: 'no', label: 'No' }] },
  { key: 'release-branch', prompt: 'Branch?', default: 'main', flag: 'release-branch', advanced: true },
  { key: 'save-recipe', prompt: 'Save?', default: 'no', flag: 'x', atEnd: true, options: [{ value: 'no', label: 'No' }] },
])

describe('parseNewQuestions', () => {
  it('reads the fields and the answers of each question', () => {
    const [ci, stack] = parseNewQuestions(JSON_TEXT)
    expect(ci).toMatchObject({ key: 'ci', flag: 'ci', many: true, advanced: false, answers: [{ value: 'github', label: 'GitHub Actions', hint: 'default' }, { value: 'azure', label: 'Azure Pipelines', hint: '' }] })
    expect(stack?.when).toEqual({ language: 'node' })
  })

  it('refuses text that is not a list', () => {
    expect(() => parseNewQuestions('{}')).toThrow('list of questions')
  })
})

describe('askableQuestions', () => {
  it('skips advanced questions, questions asked at the end and gateways without a flag', () => {
    expect(askableQuestions(parseNewQuestions(JSON_TEXT)).map((question) => question.key)).toEqual(['ci', 'node-stack'])
  })
})

describe('appliesWhen', () => {
  const stack = parseNewQuestions(JSON_TEXT)[1]

  it('asks a question only when the answers it names hold', () => {
    expect(appliesWhen(stack, { language: 'node' })).toBe(true)
    expect(appliesWhen(stack, { language: 'go' })).toBe(false)
    expect(appliesWhen(stack, {})).toBe(false)
  })
})
