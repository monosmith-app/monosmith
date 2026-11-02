import { describe, expect, it } from 'vitest'
import { askedAxes, choiceNames, choicesByAxis, hostTerms, parseChoices, parseKinds, parseStacks, roleOf } from './add-options.mapper.ts'

describe('parseKinds', () => {
  it('keeps the kinds with a name and a language', () => {
    const json = JSON.stringify([{ kind: 'go-lib', language: 'go', label: 'Go lib', description: 'a library', flags: ['stack'] }, { kind: '', language: 'go' }, 'x'])
    expect(parseKinds(json)).toEqual([{ kind: 'go-lib', language: 'go', label: 'Go lib', description: 'a library' }])
  })

  it('rejects text that is not a list', () => {
    expect(() => parseKinds('{}')).toThrow('smith printed no list')
  })
})

describe('parseChoices', () => {
  it('reads the axis, the requirements and the default', () => {
    const json = JSON.stringify([
      { axis: 'style', name: 'chakra', description: 'Chakra', language: 'node', requires: ['react'] },
      { axis: 'test', name: 'vitest', description: 'Vitest', default: true },
    ])
    expect(parseChoices(json)).toEqual([
      { axis: 'style', name: 'chakra', description: 'Chakra', language: 'node', requires: ['react'], isDefault: false, many: false, needs: [], suggested: false, conflicts: [] },
      { axis: 'test', name: 'vitest', description: 'Vitest', requires: [], isDefault: true, many: false, needs: [], suggested: false, conflicts: [] },
    ])
  })
})

describe('parseStacks', () => {
  it('reads the name, the description and the language', () => {
    expect(parseStacks(JSON.stringify([{ name: 'node-vitest', description: 'd', language: 'node', source: 'x' }]))).toEqual([{ name: 'node-vitest', description: 'd', language: 'node' }])
  })
})

describe('roleOf', () => {
  it('takes the language off a combined kind', () => {
    expect(roleOf('node-app', 'node')).toBe('app')
    expect(roleOf('python-package', 'python')).toBe('package')
    expect(roleOf('app', 'node')).toBe('app')
  })
})

describe('choicesByAxis', () => {
  const choices = parseChoices(JSON.stringify([
    { axis: 'framework', name: 'react', language: 'node' },
    { axis: 'framework', name: 'svelte', language: 'node' },
    { axis: 'test', name: 'gotest', language: 'go' },
    { axis: 'lint', name: 'any' },
  ]))

  it('groups the choices of the language by axis, and leaves the others out', () => {
    expect(choicesByAxis(choices, 'node').map((axis) => [axis.axis, axis.choices.map((choice) => choice.name)])).toEqual([['framework', ['react', 'svelte']], ['lint', ['any']]])
  })

  it('has no axis for a language without choices', () => {
    expect(choicesByAxis(choices.filter((choice) => choice.language !== undefined), 'rust')).toEqual([])
  })
})

describe('choiceNames', () => {
  const choices = parseChoices(JSON.stringify([
    { axis: 'framework', name: 'react', language: 'node' },
    { axis: 'style', name: 'react', language: 'node' },
    { axis: 'framework', name: 'fastapi', language: 'python' },
  ]))

  it('lists each name of the language once', () => {
    expect(choiceNames(choices, 'node')).toEqual(['react'])
    expect(choiceNames(choices, 'python')).toEqual(['fastapi'])
    expect(choiceNames(choices, 'rust')).toEqual([])
  })
})

describe('the axes smith asks', () => {
  const choices = parseChoices(JSON.stringify([
    { axis: 'framework', name: 'azure-functions', description: 'a serverless Azure Functions app', language: 'node', needs: ['trigger', 'runtime', 'lint'] },
    { axis: 'framework', name: 'react', description: 'a React app', language: 'node', needs: ['lint'] },
    { axis: 'lint', name: 'eslint', description: 'lint', language: 'node', default: true },
    { axis: 'trigger', name: 'http-trigger', description: 'an HTTP request', language: 'node', suggested: true },
    { axis: 'runtime', name: 'node24', description: 'Node.js 24', language: 'node', suggested: true, conflicts: ['gcp-functions'] },
    { axis: 'framework', name: 'azure-functions', description: 'a serverless Azure Functions app', language: 'python', needs: ['trigger'] },
  ]))

  it('are the axes a choice needs and no choice defaults', () => {
    expect(askedAxes(choices, 'node')).toEqual(['trigger', 'runtime'])
    expect(askedAxes(choices, 'rust')).toEqual([])
  })

  it('reads the suggested answer and the conflicts', () => {
    expect(choices.find((choice) => choice.name === 'node24')).toMatchObject({ suggested: true, conflicts: ['gcp-functions'] })
  })

  it('describe the hosts in words, so the language filter finds them', () => {
    expect(hostTerms(choices, 'node')).toEqual(['a serverless Azure Functions app'])
  })
})
