import type { SmithConnection } from '../smith-client/index.ts'
import type { RunPlan } from './run-plan.contract.ts'
import type { RunEvent, RunResult } from './run-event.contract.ts'

/** A plan run under an id the caller can cancel. */
export interface RunRequest extends RunPlan {
  runId: string
}

/**
 * Runs a plan through `smith serve` and hands every event of the run to
 * `onEvent` as it arrives. A failed run is a result with `ok: false`, not an
 * error.
 */
export async function runTasks (
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
      tasks: request.tasks,
      ...(request.projects !== undefined && request.projects.length > 0 && { projects: request.projects }),
      ...((request.affected === true) && { affected: true }),
      ...((request.bail === true) && { bail: true }),
      ...((request.noCache === true) && { noCache: true }),
      ...((request.parallel ?? 0) > 0 && { parallel: request.parallel }),
      run:   request.runId,
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
