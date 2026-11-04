import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import type { AddOptions, AddRequest } from './add-options.contract.ts'
import { parseChoices, parseKinds } from './add-options.mapper.ts'
import { fitOutcomes, parseOutcomes } from './add-outcomes.mapper.ts'
import { choicesInWords, typedItem } from './add-choice-rows.mapper.ts'
import { pickChoices, type ChoicePrompts } from './pick-choices.use-case.ts'

/** Real output of `smith choices --json`, `smith compose --json` and `smith kinds --json` (2.0.0-beta.23), not invented JSON. */
const read = (name: string): string => readFileSync(new URL(`fixtures/${name}.fixture.json`, import.meta.url), 'utf8')
const catalogue = parseOutcomes(read('compose-report'))
const options: AddOptions = {
  kinds:     parseKinds(read('kinds')),
  choices:   parseChoices(read('choices')),
  stacks:    [],
  outcomes:  fitOutcomes(catalogue.outcomes, parseKinds(read('kinds'))),
  languages: catalogue.languages,
  layers:    catalogue.layers,
}

/** A person who only presses Enter: every list takes its first row, and the lists that tick take what is ticked. */
function enterThrough (): { prompts: ChoicePrompts, asked: string[][] } {
  const asked: string[][] = []
  const prompts: ChoicePrompts = {
    many: async (question, rows) => {
      asked.push([question.axis, ...rows.map((row) => row.name)])

      return rows.filter((row) => row.picked).map((row) => row.name)
    },
    one: async (question, rows) => {
      asked.push([question.axis, ...rows.map((row) => row.name)])

      return rows[0]
    },
    asked: async (question, rows) => {
      asked.push([question.axis, ...rows.map((row) => row.name)])

      return rows[0]?.name
    },
    again: async () => false,
  }

  return { prompts, asked }
}

describe('the choices of a kind of project', () => {
  const kinded: AddOptions = {
    ...options,
    choices: [
      { axis: 'framework', name: 'aws-lambda', description: 'x', language: 'node', requires: [], isDefault: false, many: false, needs: [], suggested: false, conflicts: [], kinds: ['app'] },
      { axis: 'framework', name: 'esm', description: 'x', language: 'node', requires: [], isDefault: false, many: false, needs: [], suggested: false, conflicts: [], kinds: ['package'] },
      { axis: 'lint', name: 'eslint', description: 'x', language: 'node', requires: [], isDefault: true, many: false, needs: [], suggested: false, conflicts: [] },
    ],
  }

  it('does not list aws-lambda for a package', async () => {
    const { prompts, asked } = enterThrough()
    await pickChoices(kinded, 'node', [], undefined, prompts, 'package')
    expect(asked.find(([axis]) => axis === 'framework')?.slice(1)).toEqual(['', 'esm'])
  })

  it('says what is wrong and starts the questions again when the person agrees', async () => {
    const { prompts, asked } = enterThrough()
    const problems: string[] = []
    const again = async (problem: string): Promise<boolean> => {
      problems.push(problem)

      return true
    }
    const checks = ['this kind of project needs a framework: tick one of esm, or untick everything', '']
    const picked = await pickChoices(kinded, 'node', [], undefined, { ...prompts, again }, 'package', async () => checks.shift() ?? '')
    expect(problems).toEqual(['this kind of project needs a framework: tick one of esm, or untick everything'])
    expect(asked.filter(([axis]) => axis === 'framework')).toHaveLength(2)
    expect(picked).toBeDefined()
  })

  it('cancels when the person does not want to pick again', async () => {
    const { prompts } = enterThrough()
    expect(await pickChoices(kinded, 'node', [], undefined, prompts, 'package', async () => 'a problem')).toBeUndefined()
  })
})

describe('pressing Enter through the choices of a website', () => {
  it('gives the defaults and nothing extra: no desktop shell, no style, Vitest', async () => {
    const { prompts } = enterThrough()
    const picked = await pickChoices(options, 'node', ['react'], 'web', prompts)
    expect(picked).toEqual(['react', 'eslint', 'vitest'])
    expect(picked).not.toContain('tauri')
    expect(picked).not.toContain('bootstrap')
    expect(picked).not.toContain('jest')
  })

  it('does not ask the framework again after it was answered', async () => {
    const { prompts, asked } = enterThrough()
    await pickChoices(options, 'node', ['react'], 'web', prompts)
    expect(asked.map(([axis]) => axis)).not.toContain('framework')
  })

  it('lists a website only the frameworks of a website when the framework was left for later, and starts on "None"', async () => {
    const { prompts, asked } = enterThrough()
    const picked = await pickChoices(options, 'node', [], 'web', prompts)
    const framework = asked.find(([axis]) => axis === 'framework') ?? []
    expect(framework.slice(1)).toEqual(['', 'html', 'react', 'svelte', 'vue'])
    expect(picked).not.toContain('aws-lambda')
    expect(picked).not.toContain('electron')
  })

  it('lists "None" first for the desktop shell and the style, the default first for the test runner', async () => {
    const { prompts, asked } = enterThrough()
    await pickChoices(options, 'node', ['react'], 'web', prompts)
    const rowsOf = (axis: string): string[] => (asked.find(([name]) => name === axis) ?? []).slice(1)
    expect(rowsOf('shell')[0]).toBe('')
    expect(rowsOf('style')[0]).toBe('')
    expect(rowsOf('test')[0]).toBe('vitest')
    expect(rowsOf('test')).toContain('jest')
  })

  it('asks the trigger and the runtime of a function app once, with the suggested answer first', async () => {
    const { prompts, asked } = enterThrough()
    const picked = await pickChoices(options, 'node', ['aws-lambda'], 'function', prompts)
    expect(picked).toContain('http-trigger')
    expect(picked).toContain('node24')
    expect(asked.map(([axis]) => axis)).not.toContain('framework')
    // Jest does not go with a function host: it is not offered.
    expect((asked.find(([axis]) => axis === 'test') ?? [])).not.toContain('jest')
  })

  it('stops when the person presses Escape', async () => {
    const { prompts } = enterThrough()
    expect(await pickChoices(options, 'node', [], 'web', { ...prompts, one: async () => undefined })).toBeUndefined()
  })
})

const request = (names: string[]): AddRequest => ({ role: 'app', name: 'web', language: 'node', with: names, fresh: false })

describe('what was chosen, in words', () => {
  it('names the choices, marks the defaults and says what was left out', () => {
    expect(choicesInWords(request(['react', 'eslint', 'vitest']), options)).toBe('Your choices: React, ESLint (default), Vitest (default); no desktop shell, no style.')
  })

  it('says so when nothing extra was chosen', () => {
    expect(choicesInWords(request([]), options)).toContain('nothing extra was chosen')
  })

  it('names a stack', () => {
    expect(choicesInWords({ ...request([]), stack: 'node-vitest' }, options)).toBe('Built from the stack node-vitest.')
  })
})

describe('the list after the first question was searched', () => {
  const web = options.outcomes.find((outcome) => outcome.id === 'web')?.items.filter((item) => item.language === 'node') ?? []

  it('starts on the framework that was typed, not on the first row', () => {
    expect(typedItem(web, 'react')).toBe('react')
    expect(typedItem(web, 'vue')).toBe('vue')
  })

  it('has no start when nothing was typed or nothing matches', () => {
    expect(typedItem(web, '')).toBeUndefined()
    expect(typedItem(web, 'zzz')).toBeUndefined()
  })
})
