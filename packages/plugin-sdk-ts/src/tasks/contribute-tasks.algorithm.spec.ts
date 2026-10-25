import { describe, expect, it } from 'vitest'
import { contributeTo } from './contribute-tasks.algorithm.ts'

describe('contributeTo', () => {
  it('keys the additions by project name and leaves out a project with none', () => {
    const projects = [
      { name: 'core', root: 'libs/core', language: 'widget', kind: 'lib' as const },
      { name: 'skip', root: 'libs/skip', language: 'widget', kind: 'lib' as const },
    ]
    expect(contributeTo(projects, p => (p.name === 'skip' ? undefined : { dependsOn: { lint: ['layout-check'] } })))
      .toEqual({ projects: { core: { dependsOn: { lint: ['layout-check'] } } } })
  })
})
