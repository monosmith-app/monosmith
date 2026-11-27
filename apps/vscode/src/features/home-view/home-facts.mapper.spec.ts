import { describe, expect, it } from 'vitest'
import { toHomeFacts } from './home-facts.mapper.ts'

const workspace = {
  root:     '/repo',
  name:     'repo',
  snapshot: {
    projects: [
      { name: 'web', dir: 'apps/web', kind: 'app', language: 'TypeScript', tasks: ['build'], deps: [] },
      { name: 'core', dir: 'libs/core', kind: 'lib', language: 'Go', tasks: ['test'], deps: [] },
    ],
    affected: new Set(['web']),
  },
}

describe('toHomeFacts', () => {
  it('sums the projects, the changes, doctor, the plan and the updates of the folders', () => {
    const facts = toHomeFacts({
      workspaces: [workspace],
      folders:    1,
      reports:    [{
        findings: [
          { check: 'lock file', ok: false, fix: { description: 'sync it' } },
          { check: 'tsconfig', ok: false, warning: true },
          { check: 'git', ok: true },
        ],
      }],
      previews: [{ results: [{ project: 'kit', current: '1.0.0', next: '1.1.0', level: 'minor', commits: ['feat: a'] }, { project: 'web', next: '0.1.0', level: 'patch' }] }],
      updates:  [{ updates: [{ plugin: 'node', name: 'zod', current: '1', latest: '2', level: 'major' }] }],
      now:      0,
    })
    expect(facts).toEqual({
      projectsLoaded: true,
      projects:       2,
      languages:      ['TypeScript', 'Go'],
      affected:       1,
      health:         { failed: 1, warnings: 1, passed: 1, fixable: 1, problems: ['lock file'], warningNotes: ['tsconfig'] },
      release:        { packages: 2, name: 'kit', from: '1.0.0', to: '1.1.0', commits: 1, major: 0, minor: 1, patch: 1 },
      releaseKnown:   true,
      outdated:       1,
    })
  })

  it('leaves out what was not read yet, and is not loaded while a folder has no projects', () => {
    const facts = toHomeFacts({ workspaces: [], folders: 1, now: 0 })
    expect(facts).toEqual({ projectsLoaded: false, projects: 0, languages: [], affected: 0, releaseKnown: false })
  })

  it('tells the newest run', () => {
    const last = { id: 'r', root: '/repo', command: 'smith run test', startedAt: 0, endedAt: 9000, ok: false, tasks: [{ task: 'web:test', status: 'failed' as const }, { task: 'core:test', status: 'passed' as const }] }
    expect(toHomeFacts({ workspaces: [workspace], folders: 1, last, now: 70_000 }).last).toEqual({
      command: 'smith run test', task: 'web:test', running: false, ok: false, stopped: false, took: '9.0s', ago: '1 min ago', passed: 1, cached: 0, failed: 1,
    })
  })
})
