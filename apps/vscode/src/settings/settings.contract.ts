/** The settings of the extension, checked and with their limits applied. */
export interface MonoSmithSettings {
  /** Refresh the view by itself when files, `smith.toml` or the branch change. */
  refreshAuto:       boolean
  /** How long to wait for a burst of changes to end before refreshing. */
  refreshDebounceMs: number
  /** How many tasks a run may run at once; 0 lets smith choose. */
  parallel:          number
  /** Put the failures found in task output in the Problems panel. */
  problems:          boolean
  bail:              boolean
  noCache:           boolean
}

/** The shortest and the longest wait before a refresh. */
export const DEBOUNCE_LIMITS = { min: 100, max: 30_000, fallback: 1500 } as const
