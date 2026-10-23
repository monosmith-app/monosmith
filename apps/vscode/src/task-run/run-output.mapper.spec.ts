import { describe, expect, it } from 'vitest'
import type { RunEvent } from './run-event.contract.ts'
import { describeEvent } from './run-output.mapper.ts'

const at = (event: Partial<RunEvent> & { event: string }): RunEvent => ({ run: 'r', seq: 1, time: '', ...event })

describe('describeEvent', () => {
  it('starts with the command line', () => {
    expect(describeEvent(at({ event: 'run.started', command: 'smith run build -p api' }))).toBe('$ smith run build -p api')
  })

  it('prefixes output with its task', () => {
    expect(describeEvent(at({ event: 'task.output', task: 'api:build', line: 'ok' }))).toBe('api:build | ok')
  })

  it('says how a task ended, how long it took and why it did not run', () => {
    expect(describeEvent(at({ event: 'task.finished', task: 'api:build', status: 'passed', durationMs: 1500 }))).toBe('api:build passed in 1.5s')
    expect(describeEvent(at({ event: 'task.finished', task: 'api:test', status: 'skipped', reason: 'because api:build did not succeed' }))).toBe('api:test skipped (because api:build did not succeed)')
  })

  it('counts what the run did', () => {
    const summary = { passed: 2, failed: 0, cached: 1, skipped: 0, durationMs: 10 }
    expect(describeEvent(at({ event: 'run.finished', ok: true, summary }))).toBe('Run finished: 2 passed, 1 cached, 0 failed, 0 skipped')
    expect(describeEvent(at({ event: 'run.finished', ok: false }))).toBe('Run failed')
  })

  it('adds nothing for the events that carry no news', () => {
    expect(describeEvent(at({ event: 'task.queued', task: 'api:build' }))).toBeUndefined()
  })
})
