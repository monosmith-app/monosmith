import { describe, expect, it } from 'vitest'
import { parseChoices, parseKinds } from './add-options.mapper.ts'
import { outcomesOf, whatIs } from './add-outcomes.mapper.ts'

const kinds = parseKinds(JSON.stringify([
  { kind: 'node-app', language: 'node', label: 'Node app' }, { kind: 'node-lib', language: 'node', label: 'Node lib' },
  { kind: 'python-app', language: 'python', label: 'Python app' }, { kind: 'python-lib', language: 'python', label: 'Python lib' },
  { kind: 'go-app', language: 'go', label: 'Go app' }, { kind: 'go-package', language: 'go', label: 'Go package' },
]))
const choices = parseChoices(JSON.stringify([
  { axis: 'framework', name: 'react', title: 'React', language: 'node', aliases: ['ui'] },
  { axis: 'framework', name: 'tauri', title: 'Tauri', layer: 'shell', language: 'node', aliases: ['desktop'] },
  { axis: 'framework', name: 'azure-functions', title: 'Azure Functions', layer: 'host', language: 'node', needs: ['trigger'] },
  { axis: 'framework', name: 'azure-functions', title: 'Azure Functions', layer: 'host', language: 'python', needs: ['trigger'] },
  { axis: 'trigger', name: 'http-trigger', layer: 'host', language: 'node' },
]))

describe('whatIs', () => {
  it('says in one line what a layer means, and nothing for another word', () => {
    expect(whatIs('shell')).toBe('shell = what wraps the app as a program you install on a computer.')
    expect(whatIs('host')).toContain('where the code runs in the cloud')
    expect(whatIs('klingon')).toBe('')
  })
})

describe('outcomesOf', () => {
  const outcomes = outcomesOf({ kinds, choices, stacks: [] })
  const byId = (id: string): (typeof outcomes)[number] | undefined => outcomes.find((outcome) => outcome.id === id)

  it('takes the languages of a website, a function app and a desktop app from the choices on their layer', () => {
    expect(byId('web')?.languages).toEqual(['node'])
    expect(byId('function')?.languages).toEqual(['node', 'python'])
    expect(byId('function')?.items.map((item) => item.name)).toEqual(['azure-functions', 'azure-functions'])
    expect(byId('desktop')?.words).toEqual(expect.arrayContaining(['tauri', 'Tauri', 'desktop']))
  })

  it('keeps the backend languages for an API, in the order people look for them', () => {
    expect(byId('api')?.languages).toEqual(['node', 'python', 'go'])
  })

  it('offers a library only in the languages that have a library kind, and a package only where one exists', () => {
    expect(byId('library')?.languages).toEqual(['node', 'python'])
    expect(byId('package')?.languages).toEqual(['go'])
  })

  it('leaves out an outcome whose layer has no choice', () => {
    const none = outcomesOf({ kinds, choices: [], stacks: [] })
    expect(none.map((outcome) => outcome.id)).toEqual(['api', 'library', 'package'])
  })
})
