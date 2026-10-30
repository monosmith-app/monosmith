import type { SmithLocation, SmithLookup, SmithSession } from '../smith-client/index.ts'
import type { FolderPhase } from './folder-state.contract.ts'

/** What connecting needs from the outside. */
export interface ConnectDeps {
  locate: (root: string) => SmithLookup
  start:  (executable: string, root: string) => Promise<SmithSession>
}

/** How connecting to one folder ended. */
export interface ConnectOutcome {
  phase:            Exclude<FolderPhase, 'starting'>
  session?:         SmithSession
  location?:        SmithLocation
  pathSmith?:       string
  skippedUntrusted: boolean
  error?:           string
}

/**
 * Finds the folder's smith and starts it. It never throws: a missing smith, an
 * untrusted folder and a smith that fails to start are outcomes the view shows.
 */
export async function connectFolder (deps: ConnectDeps, root: string): Promise<ConnectOutcome> {
  const { location, pathSmith, skippedUntrusted } = deps.locate(root)
  if (location === undefined) {
    return { phase: skippedUntrusted ? 'untrusted' : 'missing', pathSmith, skippedUntrusted }
  }
  try {
    const session = await deps.start(location.executable, root)

    return { phase: 'ready', session, location, pathSmith, skippedUntrusted }
  } catch (error) {
    const reason = error instanceof Error ? error.message : String(error)

    return { phase: 'failed', location, pathSmith, skippedUntrusted, error: location.source === 'setting' ? `the monosmith.path setting (${location.executable}) did not start: ${reason}` : reason }
  }
}
