import { describe, expect, it } from 'vitest'
import { toGraphDocument } from './graph-document.mapper.ts'

const answer = {
  projects: [
    { name: 'web', kind: 'app', language: 'node', deps: ['ui', 'sdk'], marked: true },
    { name: 'ui', kind: 'lib', language: 'node', deps: ['core'] },
    { name: 'sdk', kind: 'package', language: 'node', deps: ['core'] },
    { name: 'core', kind: 'lib', language: 'node', deps: [] },
    { name: 'docs', kind: 'app', language: 'node', deps: ['web'] },
  ],
}

describe('toGraphDocument', () => {
  it('draws what the project uses down to the bottom', () => {
    const document = toGraphDocument('web', answer)
    expect(document).toContain([
      '```text',
      'web',
      '├─ ui (lib, node)',
      '│  └─ core (lib, node)',
      '└─ sdk (package, node)',
      '   └─ core (lib, node)',
      '```',
    ].join('\n'))
  })

  it('draws what uses the project', () => {
    expect(toGraphDocument('core', answer)).toContain('├─ ui (lib, node)\n│  └─ web (app, node)\n│     └─ docs (app, node)')
  })

  it('says when a side is empty', () => {
    const document = toGraphDocument('docs', answer)
    expect(document).toContain('Nothing: no other project depends on it.')
    expect(toGraphDocument('core', answer)).toContain('Nothing: it depends on no other project.')
  })

  it('does not follow a cycle', () => {
    const document = toGraphDocument('a', { projects: [{ name: 'a', deps: ['b'] }, { name: 'b', deps: ['a'] }] })
    expect(document).toContain('└─ b\n   └─ a (cycle)')
  })
})
