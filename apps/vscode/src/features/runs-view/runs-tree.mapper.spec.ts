import { describe, expect, it } from 'vitest'
import type { RunRecord } from './run-history.store.ts'
import { agoText, projectRuns, runState, testsState, toRunRows, worstStatus } from './runs-tree.mapper.ts'

const failed: RunRecord = {
  id:        'r2',
  root:      '/repo',
  command:   'smith run build test',
  startedAt: 0,
  endedAt:   12_000,
  ok:        false,
  tasks:     [
    { task: 'web:build', status: 'passed', durationMs: 4100 },
    { task: 'web:test', status: 'failed', reason: 'exit 1' },
    { task: 'core:build', status: 'cached' },
  ],
}
const running: RunRecord = { id: 'r3', root: '/repo', command: 'smith run lint', startedAt: 20_000, tasks: [{ task: 'web:lint', status: 'running' }, { task: 'core:lint', status: 'queued' }] }
const all = { errorsOnly: false }

describe('runState', () => {
  it('tells a run that goes, passed, failed or was stopped', () => {
    expect(runState(running)).toBe('running')
    expect(runState(failed)).toBe('failed')
    expect(runState({ ...failed, ok: true })).toBe('passed')
    expect(runState({ ...failed, tasks: [{ task: 'a:b', status: 'failed', reason: 'cancelled' }] })).toBe('stopped')
  })
})

describe('toRunRows', () => {
  it('says nothing ran yet before the first run', () => {
    expect(toRunRows([], 0, all)).toEqual({ nodes: [], message: 'Nothing ran yet. A run started anywhere opens here: its tasks, their marks and their logs.' })
  })

  it('opens the newest run: its state, how far it is, and its tasks grouped by project under All output', () => {
    const { nodes, summary } = toRunRows([failed], 60_000, all)
    expect(summary).toBe('smith run build test')
    expect(nodes[0]).toMatchObject({ id: 'run:r2', label: 'Failed 3 tasks', description: '3/3 100% · 12.0s', icon: 'error', color: 'charts.red', expanded: true, context: 'run.ended' })
    expect(nodes[0]?.tooltip).toContain('$ smith run build test')
    const children = nodes[0]?.children ?? []
    expect(children.map((child) => child.label)).toEqual(['All output', 'web', 'core'])
    expect(children[0]).toMatchObject({ command: 'monosmith.showOutput' })
    expect(children[1]?.children).toMatchObject([
      { label: 'build', description: '4.1s', icon: 'pass', command: 'monosmith.runs.showLog', data: 'web:build', context: 'run.task' },
      { label: 'test', description: 'exit 1', icon: 'error' },
    ])
    expect(children[2]?.children?.[0]).toMatchObject({ label: 'build', description: 'hit' })
  })

  it('keeps only the failed tasks with Errors only, and says so when none failed', () => {
    const children = toRunRows([failed], 0, { errorsOnly: true }).nodes[0]?.children ?? []
    expect(children.map((child) => child.label)).toEqual(['All output', 'web'])
    expect(children[1]?.children?.map((task) => task.label)).toEqual(['test'])
    const clean = toRunRows([{ ...failed, ok: true, tasks: [{ task: 'a:b', status: 'passed' }] }], 0, { errorsOnly: true })
    expect(clean.nodes[0]?.children?.[1]?.label).toBe('No task failed.')
  })

  it('shows the progress of a run that goes', () => {
    expect(toRunRows([running], 21_000, all).nodes[0]).toMatchObject({ label: 'Running 2 tasks', description: '0/2 0% · 1.0s', icon: 'sync~spin', context: 'run.running' })
  })

  it('puts the older runs under Runs this session, closed', () => {
    const { nodes } = toRunRows([running, failed], 30_000, all)
    expect(nodes[1]).toMatchObject({ label: 'Runs this session', description: '1' })
    expect(nodes[1]?.children?.[0]).toMatchObject({ id: 'run:r2', expanded: false })
  })
})

describe('projectRuns', () => {
  it('lists the runs that ran a task of the project, with the worst status of its tasks', () => {
    expect(projectRuns([running, failed], '/repo', 'web', 72_000)).toEqual([
      { id: 'r3', tasks: 'lint', status: 'running', note: '52.0s · just now' },
      { id: 'r2', tasks: 'build test', status: 'failed', note: '12.0s · 1 min ago' },
    ])
    expect(projectRuns([failed], '/other', 'web', 0)).toEqual([])
    expect(projectRuns([failed], '/repo', 'api', 0)).toEqual([])
  })
})

describe('testsState', () => {
  it('reads the newest run that ended with a test task', () => {
    expect(testsState([running, failed])).toBe('failed')
    expect(testsState([{ ...failed, tasks: [{ task: 'web:test', status: 'passed' }] }])).toBe('passed')
    expect(testsState([running])).toBe('none')
  })
})

describe('worstStatus and agoText', () => {
  it('put a failure first', () => {
    expect(worstStatus(['passed', 'failed', 'running'])).toBe('failed')
    expect(worstStatus([])).toBeUndefined()
  })

  it('say how long ago, in the words of the terminal', () => {
    expect(agoText(0, 30_000)).toBe('just now')
    expect(agoText(0, 180_000)).toBe('3 min ago')
    expect(agoText(0, 7_200_000)).toBe('2 h ago')
  })
})
