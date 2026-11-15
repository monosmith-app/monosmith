import type { RunEvent, TaskResult, TaskStatus } from './run-event.contract.ts'

/** What the last run did to each task, kept so the views can show it. */
export interface TaskStatusStore {
  /** Folds one event of a run into the statuses. */
  apply:       (event: RunEvent) => void
  /** The status of a task id (`api:build`), if a run has touched it. */
  statusOf:    (task: string) => TaskStatus | undefined
  /** How the task last ended or where it is: its status, duration and reason. */
  resultOf:    (task: string) => TaskResult | undefined
  /** How many tasks the runs so far left failed. */
  failedCount: () => number
  /** Listens to changes; returns what stops the listening. */
  onChange:    (listener: () => void) => () => void
}

/** A store with no statuses. */
export function createTaskStatusStore (): TaskStatusStore {
  const results = new Map<string, TaskResult>()
  const listeners = new Set<() => void>()

  const set = (task: string | undefined, result: TaskResult): void => {
    if (task === undefined) return
    results.set(task, result)
  }

  return {
    apply: (event) => {
      switch (event.event) {
        case 'run.started': {
          for (const task of listed(event)) set(task, { status: 'queued', time: event.time })
          break
        }
        case 'task.queued': {
          set(event.task, { status: 'queued', time: event.time })
          break
        }
        case 'task.started': {
          set(event.task, { status: 'running', time: event.time })
          break
        }
        case 'task.finished': {
          set(event.task, {
            status: event.status ?? 'failed',
            time:   event.time,
            ...(event.durationMs !== undefined && { durationMs: event.durationMs }),
            ...(event.reason !== undefined && { reason: event.reason }),
          })
          break
        }
        default: {
          return
        }
      }
      for (const listener of listeners) listener()
    },
    statusOf:    (task) => results.get(task)?.status,
    resultOf:    (task) => results.get(task),
    failedCount: () => {
      let failed = 0
      for (const result of results.values()) if (result.status === 'failed') failed++

      return failed
    },
    onChange: (listener) => {
      listeners.add(listener)

      return () => { listeners.delete(listener) }
    },
  }
}

function listed (event: RunEvent): string[] {
  return event.tasks ?? []
}
