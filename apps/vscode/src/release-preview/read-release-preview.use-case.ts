import type { SmithConnection } from '../smith-client/index.ts'
import type { ReleasePreview } from './release-report.contract.ts'

/** Asks smith what a release would do now (`smith release --dry-run`). It publishes nothing. */
export async function readReleasePreview (connection: SmithConnection): Promise<ReleasePreview> {
  return await connection.request<ReleasePreview>('release.preview')
}
