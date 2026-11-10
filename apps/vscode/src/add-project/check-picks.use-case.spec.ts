import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import type { AddOptions, ChoiceOption } from './add-options.contract.ts'
import { checkPicks, plainWords } from './check-picks.use-case.ts'

function choice (name: string, layer: string, over: Partial<ChoiceOption> = {}): ChoiceOption {
  return { axis: layer, name, description: '', language: 'node', requires: [], isDefault: false, many: false, needs: [], suggested: false, conflicts: [], layer, ...over }
}

const options: AddOptions = {
  kinds:     [],
  stacks:    [],
  languages: [],
  layers:    {},
  choices:   [
    choice('html', 'framework', { kinds: ['app', 'package'] }),
    choice('react', 'framework', { kinds: ['app'] }),
    choice('svelte', 'framework', { kinds: ['app'] }),
    choice('vue', 'framework', { kinds: ['app'] }),
    choice('aws-lambda', 'host', { kinds: ['app'] }),
    choice('eslint', 'lint'),
    choice('jest', 'test'),
  ],
  outcomes: [{ id: 'web', title: 'A website', role: 'app', languages: ['node'], typical: '', consequence: '', words: [], items: [{ name: 'react', title: 'React', description: '', language: 'node', layer: 'framework', aliases: [] }] }],
}

function says (check: unknown, calls: string[][] = []): (args: string[]) => Promise<SmithRun> {
  return async (args) => {
    calls.push(args)

    return { code: 0, stdout: JSON.stringify({ check }), stderr: '' }
  }
}

describe('checkPicks', () => {
  it('asks smith with the shared flags, and still asks a website for its framework', async () => {
    const calls: string[][] = []
    const problem = await checkPicks(says({ complete: true }, calls), options, { language: 'node', picks: ['eslint', 'jest'], outcomeId: 'web', role: 'app' })
    expect(calls).toEqual([['compose', '--json', '--lang', 'node', '--with', 'eslint,jest']])
    expect(problem).toBe('this kind of project needs a framework: tick one of html, react, svelte, vue, or untick everything')
  })

  it('finds lint and test choices alone fine for a package, which names no layer', async () => {
    expect(await checkPicks(says({ complete: true }), options, { language: 'node', picks: ['eslint', 'jest'], role: 'package' })).toBe('')
  })

  it('names only the choices that fit the kind', async () => {
    const problem = await checkPicks(says({ complete: true }), options, { language: 'node', picks: ['eslint'], outcomeId: 'web', role: 'package' })
    expect(problem).toBe('this kind of project needs a framework: tick one of html, or untick everything')
  })

  it('is fine when a choice of the layer is ticked', async () => {
    expect(await checkPicks(says({ complete: true }), options, { language: 'node', picks: ['react', 'eslint'], outcomeId: 'web', role: 'app' })).toBe('')
  })

  it('says a conflict and what is missing as smith words it, without flags', async () => {
    expect(await checkPicks(says({ conflict: 'react and vue are both frameworks' }), options, { language: 'node', picks: ['react', 'vue'], role: 'app' })).toBe('react and vue are both frameworks')
    expect(await checkPicks(says({ missing: 'add chakra to --with, or give no --with to start' }), options, { language: 'node', picks: ['chakra'], role: 'app' })).toBe('add chakra, or untick everything to start')
  })

  it('leaves the question of smith to the form, checks nothing for no picks, and leaves a smith that fails to smith add', async () => {
    expect(await checkPicks(says({ question: { axis: 'trigger' } }), options, { language: 'node', picks: ['aws-lambda'], outcomeId: 'web', role: 'app' })).toBe('')
    expect(await checkPicks(says({}), options, { language: 'node', picks: [] })).toBe('')
    expect(await checkPicks(async () => ({ code: 1, stdout: '', stderr: 'x' }), options, { language: 'node', picks: ['eslint'], outcomeId: 'web' })).toBe('')
  })
})

describe('plainWords', () => {
  it('drops the flag the form never types', () => {
    expect(plainWords('add x to --with')).toBe('add x')
  })
})
