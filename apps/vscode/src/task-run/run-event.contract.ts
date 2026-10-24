/** How a task ended, or where it is. */
export type TaskStatus = 'queued' | 'running' | 'passed' | 'failed' | 'cached' | 'skipped'

/** One event of a run, as `smith serve` sends it. Unknown fields are ignored. */
export interface RunEvent {
  event:       string
  run:         string
  seq:         number
  time:        string
  command?:    string
  tasks?:      string[]
  task?:       string
  stream?:     string
  line?:       string
  status?:     'passed' | 'failed' | 'cached' | 'skipped'
  durationMs?: number
  reason?:     string
  ok?:         boolean
  summary?:    { passed: number, failed: number, cached: number, skipped: number, durationMs: number }
}

/** What a finished run answers. */
export interface RunResult {
  run:      string
  ok:       boolean
  summary?: RunEvent['summary']
}

/** How the last run of a task ended: what a row of the view tells in its tooltip. */
export interface TaskResult {
  status:      TaskStatus
  /** How long it ran; absent while it has not ended, and 0 for a cached or skipped one. */
  durationMs?: number
  /** Why it was skipped or failed. */
  reason?:     string
  /** When the event came, RFC 3339. */
  time?:       string
}
