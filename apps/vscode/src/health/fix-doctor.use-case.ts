import type { SmithConnection } from '../smith-client/index.ts'
import type { FixPreview } from './fix-preview.mapper.ts'
import { readFixPreview } from './read-health.use-case.ts'

/** How an attempt to fix the workspace ended. */
export type FixOutcome = 'nothing' | 'declined' | 'fixed' | 'failed'

/**
 * Applies the fixes doctor proposes, after the person has seen them: reads the
 * preview, asks `confirm`, and only on a yes runs `doctor.fix`.
 * @param connection - the running smith
 * @param confirm - shows the fixes and answers whether to apply them
 * @param onLine - hears each line the fixing prints
 */
export async function fixDoctor (
  connection: SmithConnection,
  confirm: (preview: FixPreview) => Promise<boolean>,
  onLine: (line: string) => void = () => {},
): Promise<FixOutcome> {
  const preview = await readFixPreview(connection)
  if (preview === undefined) return 'nothing'
  if (!await confirm(preview)) return 'declined'
  const runId = `ide-fix-${Date.now()}`
  const stop = connection.onEvent((event) => {
    const seen = event as { event?: string, run?: string, line?: string }
    if (seen.run === runId && seen.event === 'task.output' && seen.line !== undefined) onLine(seen.line)
  })
  try {
    const result = await connection.request<{ ok?: boolean }>('doctor.fix', { run: runId })

    return result.ok === false ? 'failed' : 'fixed'
  } finally {
    stop()
  }
}
