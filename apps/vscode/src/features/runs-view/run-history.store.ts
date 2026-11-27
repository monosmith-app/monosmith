import type { RunEvent, TaskStatus } from '../task-run/index.ts'

/** One task of a run as the Runs view tells it. */
export interface RunTask {
  /** The task id, `web:build`. */
  task:        string
  status:      TaskStatus
  durationMs?: number
  /** Why it failed or was skipped (`cancelled` for a stopped run). */
  reason?:     string
}

/** A run started from the extension, kept for the session. */
export interface RunRecord {
  /** The id the extension gave the run. */
  id:        string
  /** The folder it ran in. */
  root:      string
  /** The smith command line it stands for, `smith run build -p web`. */
  command:   string
  /** When it started and ended, in milliseconds since the epoch; `endedAt` is absent while it runs. */
  startedAt: number
  endedAt?:  number
  /** How it ended; absent while it runs. */
  ok?:       boolean
  /** Every task, in the order smith first named it. */
  tasks:     RunTask[]
}

/** The runs of the session, the newest first. */
export interface RunHistoryStore {
  /**
   * Starts keeping a run.
   * @param root - the folder it runs in
   * @param id - the id the extension gave it
   * @param command - the command line known before smith answers
   * @param now - the time, in milliseconds
   */
  begin: (root: string, id: string, command: string, now: number) => void
  /**
   * Folds one event into its run. Output lines change nothing, so they do not redraw the view.
   * @returns true when the event changed what the view shows
   */
  apply: (event: RunEvent, now: number) => boolean
  /** Ends a run that ended without its last event (smith stopped, or the request failed). */
  end:   (id: string, ok: boolean, now: number) => void
  /** The runs, the newest first. */
  runs:  () => readonly RunRecord[]
  /** One run by its id. */
  find:  (id: string) => RunRecord | undefined
}

/**
 * Keeps the runs of the session, as the terminal interface keeps them: at most `limit`, the
 * oldest forgotten first, in memory only.
 * @param limit - how many runs to keep; the terminal interface keeps 20
 * @returns an empty history
 */
export function createRunHistory (limit = 20): RunHistoryStore {
  const runs: RunRecord[] = []
  const taskOf = (run: RunRecord, task: string): RunTask => {
    let found = run.tasks.find((item) => item.task === task)
    if (found === undefined) {
      found = { task, status: 'queued' }
      run.tasks.push(found)
    }

    return found
  }

  return {
    begin: (root, id, command, now) => {
      runs.unshift({ id, root, command, startedAt: now, tasks: [] })
      runs.splice(limit)
    },
    apply: (event, now) => {
      const run = runs.find((item) => item.id === event.run)
      if (run === undefined) return false
      switch (event.event) {
        case 'run.started': {
          if (event.command !== undefined && event.command !== '') run.command = event.command
          const named = event.tasks ?? []
          for (const task of named) taskOf(run, task)

          return true
        }
        case 'task.queued': {
          if (event.task !== undefined) taskOf(run, event.task)

          return true
        }
        case 'task.started': {
          if (event.task !== undefined) taskOf(run, event.task).status = 'running'

          return true
        }
        case 'task.finished': {
          if (event.task === undefined) return false
          const task = taskOf(run, event.task)
          task.status = event.status ?? 'failed'
          if (event.durationMs !== undefined) task.durationMs = event.durationMs
          if (event.reason !== undefined) task.reason = event.reason

          return true
        }
        case 'run.finished': {
          run.ok = event.ok === true
          run.endedAt = now

          return true
        }
        default: {
          return false
        }
      }
    },
    end: (id, ok, now) => {
      const run = runs.find((item) => item.id === id)
      if (run === undefined || run.endedAt !== undefined) return
      run.ok = ok
      run.endedAt = now
    },
    runs: () => runs,
    find: (id) => runs.find((item) => item.id === id),
  }
}
