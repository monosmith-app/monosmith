import type { SmithConnection } from '../smith-client/index.ts'
import type { UpdatesReport } from './updates-report.contract.ts'

/** Asks smith which declared dependencies and pinned tools have a newer release. It writes nothing. */
export async function readUpdates (connection: SmithConnection): Promise<UpdatesReport> {
  return await connection.request<UpdatesReport>('updates')
}
