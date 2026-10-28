import type { SmithLocation, SmithSession } from '../smith-client/index.ts'
import type { WorkspaceSnapshot } from '../workspace/index.ts'

/** Where a folder's smith is. */
export type FolderPhase = 'starting' | 'ready' | 'missing' | 'untrusted' | 'failed'

/** Everything the extension knows about one folder that has a `smith.toml`. */
export interface FolderState {
  root:             string
  name:             string
  phase:            FolderPhase
  /** Its running `smith serve`, when ready. */
  session?:         SmithSession
  /** Which smith was chosen, and why. */
  location?:        SmithLocation
  /** The smith a terminal would run. */
  pathSmith?:       string
  /** The folder's own smith was left alone because the folder is not trusted. */
  skippedUntrusted: boolean
  /** Why the folder is `failed`. */
  error?:           string
  snapshot?:        WorkspaceSnapshot
}

/**
 * What the whole view shows, from the phases of all its folders. `adoptable`:
 * no folder runs smith, but one looks like a repository smith can bring in.
 */
export type ViewState = FolderPhase | 'adoptable' | 'empty'
