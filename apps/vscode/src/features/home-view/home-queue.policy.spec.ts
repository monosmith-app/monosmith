import { describe, expect, it } from 'vitest'
import type { HomeFacts } from './home-facts.contract.ts'
import { homeQueue } from './home-queue.policy.ts'

const clean: HomeFacts = {
  projectsLoaded: true,
  projects:       4,
  languages:      ['Go'],
  affected:       0,
  health:         { failed: 0, warnings: 0, passed: 40, fixable: 0, problems: [], warningNotes: [] },
  releaseKnown:   true,
  outdated:       0,
}

describe('homeQueue', () => {
  it('offers to add the first project to an empty workspace, with the smith commands', () => {
    expect(homeQueue({ ...clean, projects: 0 })).toEqual([{
      kind:     'empty',
      severity: 'info',
      title:    'This workspace has no projects yet',
      detail:   'Say what you are building and smith picks the tools, or compose a stack from frameworks, desktop shells and cloud hosts',
      buttons:  [
        { label: 'Add a project', command: 'monosmith.addProject', line: 'smith add' },
        { label: 'Compose a stack', command: 'monosmith.composeStack', line: 'smith compose' },
      ],
    }])
  })

  it('waits for the first doctor run', () => {
    const { health: _health, ...waiting } = clean
    expect(homeQueue(waiting)).toMatchObject([{ kind: 'checking', title: 'Running the health checks on 4 projects…', buttons: [] }])
  })

  it('says all good first, with a build, when nothing asks for attention', () => {
    const queue = homeQueue({ ...clean, last: { command: 'smith run build', running: false, ok: true, stopped: false, took: '3.0s', ago: 'just now', passed: 2, cached: 0, failed: 0 } })
    expect(queue).toEqual([{ kind: 'good', severity: 'success', title: 'All good', detail: 'Last run: build passed in 3.0s', buttons: [{ label: 'Build all', command: 'monosmith.buildAll', line: 'smith run build' }] }])
  })

  it('lists what needs the person, the most urgent first', () => {
    const queue = homeQueue({
      ...clean,
      affected: 3,
      health:   { failed: 1, warnings: 2, passed: 30, fixable: 1, problems: ['lock file'], warningNotes: ['tsconfig drifts'] },
      last:     { command: 'smith run test', task: 'web:test', running: false, ok: false, stopped: false, took: '9s', ago: '4 min ago', passed: 1, cached: 0, failed: 1 },
      release:  { packages: 2, name: 'kit', from: '1.3.2', to: '1.4.0', commits: 4, major: 0, minor: 1, patch: 1 },
      outdated: 3,
    })
    expect(queue.map((item) => [item.kind, item.title])).toEqual([
      ['run', 'web:test failed'],
      ['setup', '1 check fails, 1 can be fixed now'],
      ['warnings', '2 warnings. First: tsconfig drifts'],
      ['changes', '3 projects are affected by your changes'],
      ['release', 'Ready to release: 2 packages, kit 1.3.2 → 1.4.0 (4 commits)'],
      ['updates', '3 dependencies have newer releases'],
    ])
    expect(queue[0]?.buttons).toEqual([{ label: 'See why', command: 'monosmith.runs.open' }, { label: 'Run it again', command: 'monosmith.rerunLast', line: 'smith run test' }])
    expect(queue[1]).toMatchObject({ detail: 'lock file', buttons: [{ label: 'Fix', command: 'monosmith.health.fix', line: 'smith doctor --fix' }, { label: 'Open Health', command: 'monosmith.health.focus' }] })
    expect(queue[3]?.buttons).toEqual([{ label: 'Test what changed', command: 'monosmith.testChanged', line: 'smith run lint test --affected' }])
  })

  it('says doctor could not run, with a retry', () => {
    const { health: _health, ...failing } = clean
    expect(homeQueue({ ...failing, healthError: 'smith crashed' })[0]).toMatchObject({ kind: 'doctor', severity: 'failure', detail: 'smith crashed', buttons: [{ label: 'Retry', command: 'monosmith.health.refresh', line: 'smith doctor' }] })
  })

  it('names the updates as not looked at while the registries were not asked, and leaves a stopped run out', () => {
    const { outdated: _outdated, ...unlooked } = clean
    const queue = homeQueue({ ...unlooked, last: { command: 'smith run test', running: false, ok: false, stopped: true, took: '1s', ago: '', passed: 0, cached: 0, failed: 0 } })
    expect(queue.map((item) => [item.kind, item.severity])).toEqual([['good', 'success'], ['updates', 'muted']])
    expect(queue[1]?.title).toBe('Updates not looked at yet')
  })

  it('names a first release', () => {
    expect(homeQueue({ ...clean, release: { packages: 1, name: 'kit', from: '', to: '0.1.0', commits: 0, major: 0, minor: 1, patch: 0 } })[0]?.title).toBe('Ready to release: 1 package, kit 0.1.0, its first release')
  })
})
