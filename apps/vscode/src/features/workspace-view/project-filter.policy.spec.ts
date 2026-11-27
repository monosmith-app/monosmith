import { describe, expect, it } from 'vitest'
import { filterProject, matchesWord } from './project-filter.policy.ts'

const web = { name: 'web', kind: 'app', language: 'TypeScript', tasks: ['build', 'test', 'lint'], affected: true }
const core = { name: 'core', kind: 'lib', language: 'Go', tasks: ['build', 'test'], affected: false }

describe('matchesWord', () => {
  it('reads kind: and lang: as prefixes', () => {
    expect(matchesWord(web, 'kind:ap')).toBe(true)
    expect(matchesWord(core, 'kind:ap')).toBe(false)
    expect(matchesWord(web, 'lang:type')).toBe(true)
    expect(matchesWord(web, 'lang:script')).toBe(false)
  })

  it('reads changed as a project the changes affect, else as text', () => {
    expect(matchesWord(web, 'changed')).toBe(true)
    expect(matchesWord(core, 'changed')).toBe(false)
  })

  it('matches any part of the name, the kind or the language', () => {
    expect(matchesWord(core, 'or')).toBe(true)
    expect(matchesWord(core, 'li')).toBe(true)
    expect(matchesWord(core, 'go')).toBe(true)
  })
})

describe('filterProject', () => {
  it('keeps every task when the words match the project', () => {
    expect(filterProject(web, 'changed kind:app')).toEqual({ tasks: ['build', 'test', 'lint'], reveal: false })
    expect(filterProject(web, '')).toEqual({ tasks: ['build', 'test', 'lint'], reveal: false })
  })

  it('keeps the tasks a word names, and opens the project on them', () => {
    expect(filterProject(web, 'lin')).toEqual({ tasks: ['lint'], reveal: true })
    expect(filterProject(web, 'kind:app lint')).toEqual({ tasks: ['lint'], reveal: true })
  })

  it('drops a project a word does not match', () => {
    expect(filterProject(core, 'changed')).toBeUndefined()
    expect(filterProject(core, 'kind:app build')).toBeUndefined()
    expect(filterProject(core, 'deploy')).toBeUndefined()
  })
})
