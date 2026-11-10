import { describe, expect, it } from 'vitest'
import { linkableLibraries, mayLink } from './linkable-libraries.policy.ts'

const nodes = [
  { name: 'web', kind: 'app', language: 'node', deps: ['ui'] },
  { name: 'ui', kind: 'lib', language: 'node', deps: ['core'] },
  { name: 'core', kind: 'lib', language: 'node', deps: [] },
  { name: 'core-py', kind: 'lib', language: 'python', deps: [] },
]

describe('mayLink', () => {
  it('refuses itself, an app, another language and a cycle', () => {
    expect(mayLink(nodes, 'ui', 'ui')).toBe(false)
    expect(mayLink(nodes, 'core', 'web')).toBe(false)
    expect(mayLink(nodes, 'web', 'core-py')).toBe(false)
    expect(mayLink(nodes, 'core', 'ui')).toBe(false)
  })

  it('accepts a library of the same language that does not depend on the project', () => {
    expect(mayLink(nodes, 'web', 'core')).toBe(true)
    expect(mayLink(nodes, 'ui', 'core')).toBe(true)
  })
})

describe('linkableLibraries', () => {
  it('lists only what smith would accept', () => {
    expect(linkableLibraries(nodes, 'core').map((node) => node.name)).toEqual([])
    expect(linkableLibraries(nodes, 'web').map((node) => node.name)).toEqual(['ui', 'core'])
  })
})
