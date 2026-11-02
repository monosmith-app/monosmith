import { describe, expect, it } from 'vitest'
import { askCustomAnswers, type PickAnswer } from './custom-answers.use-case.ts'
import { DEFAULT_CHOICES } from './new-workspace-arguments.mapper.ts'
import type { NewQuestion } from './new-workspace.contract.ts'

const ask = (key: string, values: string[], over: Partial<NewQuestion> = {}): NewQuestion => ({
  key,
  prompt:   `prompt of ${key}`,
  default:  values[0] ?? '',
  flag:     key,
  many:     false,
  advanced: false,
  atEnd:    false,
  when:     {},
  answers:  values.map((value) => ({ value, label: value, hint: '' })),
  ...over,
})

const QUESTIONS: NewQuestion[] = [
  ask('ci', ['github', 'azure'], { many: true }),
  ask('language', ['later', 'node', 'go']),
  ask('node-stack', ['node-jest', 'node-vitest'], { when: { language: 'node' } }),
  ask('architecture', ['vertical-slices', 'none']),
  ask('agent', ['yes', 'no']),
  ask('devcontainer', ['yes', 'no']),
  ask('more', ['no', 'yes'], { flag: '' }),
  ask('release-branch', ['main'], { advanced: true }),
  ask('save-recipe', ['no', 'yes'], { atEnd: true }),
]
const START = { name: 'ws', ...DEFAULT_CHOICES }

function fakeClient (script: Map<string, string[] | undefined>) {
  const asked: string[] = []
  const pick: PickAnswer = async (question, current) => {
    asked.push(`${question.key}:${current.join(',')}`)

    return script.has(question.key) ? script.get(question.key) : [...current]
  }

  return { asked, pick }
}

describe('askCustomAnswers', () => {
  it('asks the questions in smith\'s order, starting each on the current choice, and skips the rest', async () => {
    const client = fakeClient(new Map())
    expect(await askCustomAnswers(QUESTIONS, START, client.pick)).toEqual(START)
    expect(client.asked).toEqual(['ci:github', 'language:later', 'architecture:vertical-slices', 'agent:yes', 'devcontainer:yes'])
  })

  it('asks the Node stack only when the language is node, and maps the answers to the choices', async () => {
    const client = fakeClient(new Map(Object.entries({ 'ci': ['github', 'azure'], 'language': ['node'], 'node-stack': ['node-vitest'], 'architecture': ['none'], 'agent': ['no'], 'devcontainer': ['no'] })))
    expect(await askCustomAnswers(QUESTIONS, START, client.pick)).toEqual({
      name: 'ws', ci: ['github', 'azure'], language: 'node', nodeStack: 'node-vitest', architecture: 'none', agent: false, devcontainer: false, commit: true,
    })
    expect(client.asked).toContain('node-stack:node-jest')
  })

  it('stops at the first question the person cancels', async () => {
    const client = fakeClient(new Map([['language', undefined]]))
    expect(await askCustomAnswers(QUESTIONS, START, client.pick)).toBeUndefined()
    expect(client.asked).toEqual(['ci:github', 'language:later'])
  })

  it('asks a question that smith adds later without a change in the extension', async () => {
    const client = fakeClient(new Map([['shell', ['zsh']]]))
    await askCustomAnswers([...QUESTIONS, ask('shell', ['bash', 'zsh'])], START, client.pick)
    expect(client.asked.at(-1)).toBe('shell:bash')
  })
})
