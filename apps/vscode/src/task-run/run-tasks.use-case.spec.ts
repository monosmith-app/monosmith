import { describe, expect, it } from 'vitest'
import type { SmithConnection } from '../smith-client/index.ts'
import type { RunEvent } from './run-event.contract.ts'
import { cancelRun, runTasks } from './run-tasks.use-case.ts'

function fakeSmith () {
  const listeners = new Set<(event: unknown) => void>()
  const calls: Array<{ method: string, params: unknown }> = []
  const connection: SmithConnection = {
    request: async <T>(method: string, params?: unknown): Promise<T> => {
      calls.push({ method, params })
      if (method === 'run') {
        for (const listener of listeners) {
          listener({ event: 'task.started', run: 'ide-1', seq: 1, time: '', task: 'api:build' })
          listener({ event: 'task.started', run: 'someone-elses', seq: 1, time: '', task: 'web:build' })
        }

        return { run: 'ide-1', ok: true } as T
      }

      return { cancelled: true } as T
    },
    onEvent: (handler) => {
      listeners.add(handler)

      return () => { listeners.delete(handler) }
    },
    close: () => {},
  }

  return { connection, calls, listeners }
}

describe('runTasks', () => {
  it('asks smith to run the task of the project under the given id', async () => {
    const { connection, calls } = fakeSmith()
    await runTasks(connection, { runId: 'ide-1', tasks: ['build'], projects: ['api'] }, () => {})
    expect(calls[0]).toEqual({ method: 'run', params: { tasks: ['build'], projects: ['api'], run: 'ide-1' } })
  })

  it('hands over the events of its run only, and stops listening after', async () => {
    const { connection, listeners } = fakeSmith()
    const seen: RunEvent[] = []
    const result = await runTasks(connection, { runId: 'ide-1', tasks: ['build'], projects: ['api'] }, (event) => { seen.push(event) })
    expect(result.ok).toBe(true)
    expect(seen.map((event) => event.task)).toEqual(['api:build'])
    expect(listeners.size).toBe(0)
  })
})

describe('cancelRun', () => {
  it('says whether smith stopped the run', async () => {
    const { connection, calls } = fakeSmith()
    await expect(cancelRun(connection, 'ide-1')).resolves.toBe(true)
    expect(calls[0]).toEqual({ method: 'cancel', params: { run: 'ide-1' } })
  })
})
