/** The last log of each task, kept so a row can open it, also for a task that failed. */
export interface TaskLogStore {
  /** Starts the log of a task over: its last run is replaced by this one. */
  begin: (key: string, header?: string) => void
  /** Adds a line to the log of a task that began. */
  add:   (key: string, line: string) => void
  /** The lines of the last run of a task, or undefined when none ran. */
  lines: (key: string) => readonly string[] | undefined
  /** Whether a log is kept for the task. */
  has:   (key: string) => boolean
  /** Forgets every log. */
  clear: () => void
}

/**
 * A store of logs, each capped to its newest `capacity` lines so a noisy task
 * cannot grow the editor's memory without end.
 * @param capacity - the most lines kept for a task
 */
export function createTaskLogStore (capacity = 5000): TaskLogStore {
  const logs = new Map<string, string[]>()

  return {
    begin: (key, header) => { logs.set(key, header === undefined ? [] : [header]) },
    add:   (key, line) => {
      const log = logs.get(key)
      if (log === undefined) return
      log.push(line)
      if (log.length > capacity) log.splice(0, log.length - capacity)
    },
    lines: (key) => logs.get(key),
    has:   (key) => logs.has(key),
    clear: () => { logs.clear() },
  }
}
