import { describe, expect, it } from 'vitest'
import { addArguments, checkProjectName, followsSiblings } from './add-arguments.mapper.ts'

const request = { role: 'app', name: 'web', language: 'node', with: [], fresh: false }

describe('addArguments', () => {
  it('plans with --dry-run and applies with --yes', () => {
    expect(addArguments(request, 'plan')).toEqual(['add', 'app', 'web', '--lang', 'node', '--dry-run'])
    expect(addArguments(request, 'apply')).toEqual(['add', 'app', 'web', '--lang', 'node', '--yes'])
  })

  it('passes the stack, the choices and --fresh', () => {
    expect(addArguments({ ...request, stack: 'node-vitest', with: ['react', 'vitest'], fresh: true }, 'plan')).toEqual([
      'add', 'app', 'web', '--lang', 'node', '--stack', 'node-vitest', '--with', 'react,vitest', '--fresh', '--dry-run',
    ])
  })

  it('copies one project with --like, unless the stack defaults are asked for', () => {
    expect(addArguments({ ...request, like: 'api' }, 'plan')).toEqual(['add', 'app', 'web', '--lang', 'node', '--like', 'api', '--dry-run'])
    expect(addArguments({ ...request, like: 'api' }, 'plan', true)).not.toContain('--like')
  })

  it('lets the caller override --fresh', () => {
    expect(addArguments(request, 'apply', true)).toContain('--fresh')
  })
})

describe('followsSiblings', () => {
  it('sees the line smith prints for a project that follows its siblings', () => {
    expect(followsSiblings('    create apps/web (follows api: vitest)')).toBe(true)
    expect(followsSiblings('The 3 sibling packages agree on:')).toBe(true)
    expect(followsSiblings('    create apps/web')).toBe(false)
  })
})

describe('checkProjectName', () => {
  it('accepts kebab-case names', () => {
    expect(checkProjectName('billing-api')).toBeUndefined()
  })

  it('rejects a space, a slash, a capital or nothing, and says what to do', () => {
    for (const name of ['my app', 'apps/web', 'Web', '', '-web']) expect(checkProjectName(name)).toContain('Use lowercase letters')
  })
})
