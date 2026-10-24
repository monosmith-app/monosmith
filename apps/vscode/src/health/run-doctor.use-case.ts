import type { SmithConnection } from '../smith-client/index.ts'
import type { DoctorReport } from './doctor-report.contract.ts'

/** Asks smith to check the workspace's health. */
export async function runDoctor (connection: SmithConnection): Promise<DoctorReport> {
  return await connection.request<DoctorReport>('doctor')
}
