import type { RunEvent, TaskStatus } from './run-event.contract.ts'

/** What the last run did to each task, kept so the views can show it. */
export interface TaskStatusStore {
  /** Folds one event of a run into the statuses. */
  apply:    (event: RunEvent) => void
  /** The status of a task id (`api:build`), if a run has touched it. */
  statusOf: (task: string) => TaskStatus | undefined
  /** Listens to changes; returns what stops the listening. */
  onChange: (listener: () => void) => () => void
}

/** A store with no statuses. */
export function createTaskStatusStore (): TaskStatusStore {
  const statuses = new Map<string, TaskStatus>()
  const listeners = new Set<() => void>()

  const set = (task: string | undefined, status: TaskStatus): void => {
    if (task === undefined) return
    statuses.set(task, status)
  }

  return {
    apply: (event) => {
      switch (event.event) {
        case 'run.started': {
          for (const task of listed(event)) set(task, 'queued')
          break
        }
        case 'task.queued': {
          set(event.task, 'queued')
          break
        }
        case 'task.started': {
          set(event.task, 'running')
          break
        }
        case 'task.finished': {
          set(event.task, event.status ?? 'failed')
          break
        }
        default: {
          return
        }
      }
      for (const listener of listeners) listener()
    },
    statusOf: (task) => statuses.get(task),
    onChange: (listener) => {
      listeners.add(listener)

      return () => { listeners.delete(listener) }
    },
  }
}

function listed (event: RunEvent): string[] {
  return event.tasks ?? []
}
