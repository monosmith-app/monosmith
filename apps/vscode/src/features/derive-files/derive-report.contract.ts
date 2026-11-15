/** What `smith derive` finds for a file that needs action. */
export type DeriveStatus = 'stale' | 'missing' | 'unrecorded' | 'drift' | 'diverged' | 'broken'

/** One derived file that needs action, as a line of `smith derive --check`. */
export interface DerivedFile {
  /** Relative to the workspace folder. */
  path:     string
  status:   DeriveStatus
  /** The renderer that writes it, such as `tsconfig-references`. */
  renderer: string
  /** Why the file is in that state, in smith's words. */
  detail:   string
}

/**
 * The derive state of a workspace folder: `off` (`[derive] enabled` is not
 * set), `current` (nothing to do), `pending` (files need action), or `failed`
 * (smith could not check; `message` says why and the fix).
 */
export interface DeriveReport {
  state:    'off' | 'current' | 'pending' | 'failed'
  files:    DerivedFile[]
  /** How many files smith derives, those that are current included. */
  total:    number
  message?: string
}

/** What a person can do with a derived file. */
export type DeriveAction = 'check' | 'explain' | 'apply' | 'replace' | 'adopt' | 'eject'
