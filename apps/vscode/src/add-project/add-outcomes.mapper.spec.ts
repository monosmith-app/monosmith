import { describe, expect, it } from 'vitest'
import { parseKinds } from './add-options.mapper.ts'
import { fitOutcomes, languageRows, outcomeRows, parseOutcomes, whatIs, type OutcomeOption } from './add-outcomes.mapper.ts'

const kinds = parseKinds(JSON.stringify([
  { kind: 'node-app', language: 'node', label: 'Node app' }, { kind: 'node-lib', language: 'node', label: 'Node lib' },
  { kind: 'python-app', language: 'python', label: 'Python app' }, { kind: 'python-lib', language: 'python', label: 'Python lib' },
  { kind: 'go-app', language: 'go', label: 'Go app' }, { kind: 'go-package', language: 'go', label: 'Go package' },
]))

const json = JSON.stringify({
  outcomes: [
    { id: 'web', title: 'A website or web app', kind: 'app', layer: 'framework', typical: ['React', 'Vue'], consequence: 'People open it in a browser.', languages: ['node'], words: ['website', 'frontend'], items: [{ name: 'react', title: 'React', language: 'node', layer: 'framework', aliases: ['ui'] }] },
    { id: 'api', title: 'An API or backend service', kind: 'app', typical: ['Go'], consequence: 'Answers requests.', languages: ['node', 'python', 'go'], words: ['api', 'backend'] },
    { id: 'function', title: 'A function app (Azure, AWS, Google)', kind: 'app', layer: 'host', typical: ['Azure Functions'], consequence: 'The cloud runs it.', languages: ['dotnet', 'node', 'python'], words: ['serverless', 'lambda'], items: [{ name: 'azure-functions', title: 'Azure Functions', language: 'node', layer: 'host' }, { name: 'azure-functions', title: 'Azure Functions', language: 'dotnet', layer: 'host' }] },
    { id: 'desktop', title: 'A desktop app (Tauri, Electron)', kind: 'app', layer: 'shell', typical: ['Electron', 'Tauri'], consequence: 'Installed on a computer.', languages: ['node'], words: ['desktop'], items: [{ name: 'tauri', title: 'Tauri', language: 'node', layer: 'shell', aliases: ['rust shell'] }, { name: 'electron', title: 'Electron', language: 'node', layer: 'shell' }] },
    { id: 'library', title: 'A library for my other projects', kind: 'lib', typical: [], consequence: 'Shared code.', languages: ['node', 'python', 'go'], words: ['library', 'shared'] },
    { id: 'package', title: 'A package I publish', kind: 'package', typical: ['npm'], consequence: 'Published.', languages: ['node', 'go'], words: ['package', 'publish'] },
  ],
  items: [{ kind: 'language', name: 'dotnet', title: 'C# / .NET', description: 'C# apps' }, { kind: 'choice', name: 'react', title: 'React' }],
})

describe('whatIs', () => {
  it('says in one line what a layer means, and nothing for another word', () => {
    expect(whatIs('shell')).toBe('shell = what wraps the app as a program you install on a computer.')
    expect(whatIs('host')).toContain('where the code runs in the cloud')
    expect(whatIs('klingon')).toBe('')
  })
})

describe('parseOutcomes', () => {
  const read = parseOutcomes(json)

  it('reads the six outcomes of smith, in its order, with the website and the desktop app', () => {
    expect(read.outcomes.map((outcome) => outcome.id)).toEqual(['web', 'api', 'function', 'desktop', 'library', 'package'])
    expect(read.outcomes[3]?.items.map((item) => item.name)).toEqual(['tauri', 'electron'])
    expect(read.outcomes[0]?.typical).toBe('React, Vue')
  })

  it('reads the languages with the name a person knows', () => {
    expect(read.languages).toEqual([{ name: 'dotnet', title: 'C# / .NET', description: 'C# apps' }])
  })

  it('gives nothing, not an error, when smith lists no outcomes', () => {
    expect(parseOutcomes('{"items":[]}')).toEqual({ outcomes: [], languages: [] })
  })

  it('rejects text that is not a catalogue', () => {
    expect(() => parseOutcomes('[]')).toThrow('smith printed no catalogue')
  })
})

describe('fitOutcomes', () => {
  const fit = fitOutcomes(parseOutcomes(json).outcomes, kinds)
  const byId = (id: string): OutcomeOption | undefined => fit.find((outcome) => outcome.id === id)

  it('keeps the languages that have a kind of the role of the outcome', () => {
    expect(byId('library')?.languages).toEqual(['node', 'python'])
    expect(byId('package')?.languages).toEqual(['go'])
  })

  it('drops a language smith cannot add, and the items of it', () => {
    expect(byId('function')?.languages).toEqual(['node', 'python'])
    expect(byId('function')?.items.map((item) => item.language)).toEqual(['node'])
  })
})

describe('languageRows', () => {
  const read = parseOutcomes(json)
  const outcomes = fitOutcomes(read.outcomes, kinds)
  const fn = outcomes.find((outcome) => outcome.id === 'function')

  it('names the language and the clouds it offers, not a list of ids', () => {
    const rows = languageRows(read.languages, kinds, 'app', fn?.languages ?? [], fn)
    expect(rows.map((row) => row.language)).toEqual(['node', 'python'])
    expect(rows[0]).toMatchObject({ label: 'node', description: 'Azure Functions' })
    expect(rows.map((row) => row.description).join(' ')).not.toMatch(/trigger|dotnet\d/)
  })

  it('uses the title of the language when smith lists one', () => {
    const rows = languageRows([{ name: 'go', title: 'Go', description: 'Go apps' }], kinds, 'app', ['go'])
    expect(rows).toEqual([{ language: 'go', label: 'Go', description: '', detail: 'Go apps' }])
  })

  it('leaves out a language that has no kind of the role', () => {
    expect(languageRows([], kinds, 'lib', ['go', 'node'])).toHaveLength(1)
  })
})

describe('outcomeRows', () => {
  const outcomes = fitOutcomes(parseOutcomes(json).outcomes, kinds)
  const ids = (query: string): string[] => outcomeRows(outcomes, query).map((row) => (row.kind === 'outcome' ? row.outcome.id : row.kind))

  it('lists every outcome, then the long way and the composer, for an empty search', () => {
    expect(ids('')).toEqual(['web', 'api', 'function', 'desktop', 'library', 'package', 'every', 'compose'])
  })

  it('finds a website by react, a desktop app by tauri or electron, and a function app by azure', () => {
    expect(ids('react')[0]).toBe('web')
    expect(ids('tauri')[0]).toBe('desktop')
    expect(ids('electron')[0]).toBe('desktop')
    expect(ids('azure function')[0]).toBe('function')
  })

  it('finds by an alias of an item and by a word of the outcome', () => {
    expect(ids('rust shell')[0]).toBe('desktop')
    expect(ids('serverless')[0]).toBe('function')
  })

  it('says there is no result and names the closest outcomes, and keeps the ways on', () => {
    const rows = outcomeRows(outcomes, 'reakt')
    expect(rows[0]).toMatchObject({ kind: 'none', query: 'reakt' })
    expect(rows[0]?.kind === 'none' ? rows[0].closest : []).toContain('A website or web app')
    expect(rows.slice(1).map((row) => row.kind)).toEqual(['every', 'compose'])
  })
})
