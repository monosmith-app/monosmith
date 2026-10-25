import { describe, expect, it } from 'vitest'
import { projectKindDirectory, resolveProjectKind } from './project-kind.policy.ts'

describe('resolveProjectKind', () => {
  it('takes the declared kind when it is valid, trimming it', () => {
    expect(resolveProjectKind(' app ', 'libs/x', 'lib')).toBe('app')
    expect(resolveProjectKind('package', 'x', 'lib')).toBe('package')
  })

  it('falls back to the directory convention, then to the fallback', () => {
    expect(resolveProjectKind('nonsense', 'apps/web', 'lib')).toBe('app')
    expect(resolveProjectKind(undefined, 'libs/core', 'app')).toBe('lib')
    expect(resolveProjectKind('', 'packages/sdk', 'app')).toBe('package')
    expect(resolveProjectKind('', 'tools/x', 'lib')).toBe('lib')
    expect(resolveProjectKind('', 'toString', 'lib')).toBe('lib')
  })
})

describe('projectKindDirectory', () => {
  it('names the conventional parent', () => {
    expect([projectKindDirectory('app'), projectKindDirectory('lib'), projectKindDirectory('package')]).toEqual(['apps', 'libs', 'packages'])
  })
})
