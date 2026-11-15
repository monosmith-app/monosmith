import type { SmithConnection } from '../smith-client/index.ts'
import type { DoctorReport } from './doctor-report.contract.ts'
import { parseFixPreview, type FixPreview } from './fix-preview.mapper.ts'
import { runDoctor } from './run-doctor.use-case.ts'

/** The report of the doctor and the fixes it proposes for it. */
export interface HealthReading {
  report:   DoctorReport
  /** Absent when nothing failed or doctor cannot fix what did. */
  preview?: FixPreview
}

/**
 * Asks smith for the health of the workspace and, when a check failed, for
 * the fixes it would apply (a preview; nothing is written).
 */
export async function readHealth (connection: SmithConnection): Promise<HealthReading> {
  const report = await runDoctor(connection)
  const failing = (report.findings ?? []).some((finding) => !finding.ok && finding.warning !== true)

  return { report, ...(failing && { preview: await readFixPreview(connection) }) }
}

/** The fixes smith proposes, or undefined when it proposes none. */
export async function readFixPreview (connection: SmithConnection): Promise<FixPreview | undefined> {
  try {
    await connection.request<unknown>('doctor.fixes')

    return undefined
  } catch (error) {
    return parseFixPreview(error instanceof Error ? error.message : String(error))
  }
}
