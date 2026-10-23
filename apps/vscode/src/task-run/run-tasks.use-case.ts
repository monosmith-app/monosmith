import type { SmithConnection } from '../smith-client/index.ts'
import type { RunEvent, RunResult } from './run-event.contract.ts'

/** A task of one project, run under an id the caller can cancel. */
export interface RunRequest {
  runId:   string
  task:    string
  project: string
}

/**
 * Runs one task of one project through `smith serve` and hands every event of
 * the run to `onEvent` as it arrives. A failed run is a result with `ok:
 * false`, not an error.
 */
export async function runTask (
  connection: SmithConnection,
  request: RunRequest,
  onEvent: (event: RunEvent) => void,
): Promise<RunResult> {
  const stop = connection.onEvent((event) => {
    const seen = event as RunEvent
    if (seen.run === request.runId) onEvent(seen)
  })
  try {
    return await connection.request<RunResult>('run', {
      tasks:    [request.task],
      projects: [request.project],
      run:      request.runId,
    })
  } finally {
    stop()
  }
}

/** Stops a run and its processes. */
export async function cancelRun (connection: SmithConnection, runId: string): Promise<boolean> {
  const answer = await connection.request<{ cancelled: boolean }>('cancel', { run: runId })

  return answer.cancelled
}
