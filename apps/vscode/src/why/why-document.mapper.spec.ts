import { describe, expect, it } from 'vitest'
import { toFileWhyDocument, toTaskWhyDocument } from './why-document.mapper.ts'

describe('toFileWhyDocument', () => {
  const derived = JSON.stringify({
    path:     'packages/web/tsconfig.lib.json',
    known:    true,
    kind:     'derived',
    owner:    'smith',
    summary:  'smith owns this file.',
    renderer: 'tsconfig-references',
    change:   ['edit the input and run smith derive', 'or run smith derive --eject <file> to own the file'],
  })

  it('says who owns the file, the facts, and how to change it', () => {
    const page = toFileWhyDocument(derived)
    expect(page).toContain('# Why: packages/web/tsconfig.lib.json')
    expect(page).toContain('smith owns this file.')
    expect(page).toContain('- **owner:** smith')
    expect(page).toContain('- **renderer:** tsconfig-references')
    expect(page).toContain('## How to change it\n\n- edit the input and run smith derive\n- or run smith derive --eject <file> to own the file')
  })

  it('puts the owner before the other facts', () => {
    const page = toFileWhyDocument(derived)
    expect(page.indexOf('**owner:**')).toBeLessThan(page.indexOf('**renderer:**'))
  })

  it('lists no facts for a file smith does not know', () => {
    const page = toFileWhyDocument(JSON.stringify({ path: 'x.txt', known: false, kind: 'unknown', summary: 'smith does not know this file.', change: ['edit it as you like'] }))
    expect(page).not.toContain('**kind:**')
    expect(page).toContain('- edit it as you like')
  })

  it('rejects text that is not an answer', () => {
    expect(() => toFileWhyDocument('[]')).toThrow('smith printed no answer')
  })
})

describe('toTaskWhyDocument', () => {
  it('lists what changed since the last run', () => {
    const page = toTaskWhyDocument(JSON.stringify({
      task:        'api:build',
      status:      'changed',
      key:         'k',
      lastRun:     '2026-11-01T07:32:13Z',
      differences: [{ what: 'file', name: 'apps/api/new.txt', change: 'added' }],
    }))
    expect(page).toContain('# Why: api:build')
    expect(page).toContain('inputs changed')
    expect(page).toContain('- file `apps/api/new.txt`: added')
    expect(page).toContain('- **last run:** 2026-11-01T07:32:13Z')
  })

  it('says a hit comes from the cache', () => {
    expect(toTaskWhyDocument(JSON.stringify({ task: 'a:t', status: 'hit', key: 'k' }))).toContain('comes from the cache')
  })

  it('names a state it does not know', () => {
    expect(toTaskWhyDocument(JSON.stringify({ task: 'a:t', status: 'odd' }))).toContain('"odd"')
  })
})
