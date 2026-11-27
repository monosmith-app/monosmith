import { describe, expect, it } from 'vitest'
import { createRunHistory } from './run-history.store.ts'

const event = (run: string, name: string, more: Record<string, unknown> = {}) => ({ event: name, run, seq: 0, time: '', ...more })

describe('createRunHistory', () => {
  it('keeps a run from its start to its end, the tasks in the order smith named them', () => {
    const history = createRunHistory()
    history.begin('/repo', 'r1', 'smith run build', 1000)
    expect(history.apply(event('r1', 'run.started', { command: 'smith run build test -p web', tasks: ['web:build', 'web:test'] }), 1000)).toBe(true)
    history.apply(event('r1', 'task.started', { task: 'web:build' }), 1100)
    expect(history.apply(event('r1', 'task.output', { task: 'web:build', line: 'ok' }), 1150)).toBe(false)
    history.apply(event('r1', 'task.finished', { task: 'web:build', status: 'passed', durationMs: 900 }), 1200)
    history.apply(event('r1', 'task.finished', { task: 'web:test', status: 'failed', reason: 'exit 1' }), 1300)
    history.apply(event('r1', 'run.finished', { ok: false }), 1400)
    expect(history.find('r1')).toEqual({
      id:        'r1',
      root:      '/repo',
      command:   'smith run build test -p web',
      startedAt: 1000,
      endedAt:   1400,
      ok:        false,
      tasks:     [{ task: 'web:build', status: 'passed', durationMs: 900 }, { task: 'web:test', status: 'failed', reason: 'exit 1' }],
    })
  })

  it('ignores the events of a run it does not keep', () => {
    const history = createRunHistory()
    expect(history.apply(event('other', 'task.started', { task: 'a:b' }), 0)).toBe(false)
  })

  it('keeps the newest first, and forgets the oldest past its limit', () => {
    const history = createRunHistory(2)
    history.begin('/r', 'a', 'smith run a', 1)
    history.begin('/r', 'b', 'smith run b', 2)
    history.begin('/r', 'c', 'smith run c', 3)
    expect(history.runs().map((run) => run.id)).toEqual(['c', 'b'])
  })

  it('ends a run that ended without its last event, once', () => {
    const history = createRunHistory()
    history.begin('/r', 'a', 'smith run a', 1)
    history.end('a', false, 5)
    history.end('a', true, 9)
    expect(history.find('a')).toMatchObject({ ok: false, endedAt: 5 })
  })
})
