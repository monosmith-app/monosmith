import { describe, expect, it } from 'vitest'
import { mainLanguage, startOnMainLanguage } from './main-language.mapper.ts'

describe('mainLanguage', () => {
  it('is the language most projects use, with the line that says why', () => {
    const projects = [{ language: 'python' }, { language: 'node' }, { language: 'node' }]
    expect(mainLanguage(projects, ['python', 'node', 'go'])).toEqual({ language: 'node', why: 'Preselected: 2 of the 3 projects here use node.' })
  })

  it('counts only the languages that fit what is being built, and a tie goes to the first of the list', () => {
    expect(mainLanguage([{ language: 'go' }, { language: 'node' }], ['python', 'node'])?.language).toBe('node')
    expect(mainLanguage([{ language: 'go' }, { language: 'node' }], ['node', 'go'])?.language).toBe('node')
    expect(mainLanguage([{ language: 'go' }], ['python'])).toBeUndefined()
    expect(mainLanguage([], ['python'])).toBeUndefined()
  })
})

describe('startOnMainLanguage', () => {
  it('moves the main language first and keeps the others in order', () => {
    const rows = [{ language: 'python' }, { language: 'node' }, { language: 'go' }]
    expect(startOnMainLanguage(rows, { language: 'go', why: '' }).map((r) => r.language)).toEqual(['go', 'python', 'node'])
    expect(startOnMainLanguage(rows, undefined)).toEqual(rows)
  })
})
