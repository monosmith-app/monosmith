import type { FolderState } from './folder-state.contract.ts'

/** The folders with a `smith.toml` the extension knows, by root. */
export interface FolderRegistry {
  /** The state of each folder, by root: the store owns it and the callers change it in place. */
  folders: Map<string, FolderState>
  /** Every folder, ready or not. */
  states:  () => FolderState[]
  /** The folders that run smith now. */
  ready:   () => FolderState[]
  /** Ends the `smith serve` of one folder and forgets the session. */
  stop:    (root: string) => void
}

/**
 * Creates an empty registry of folders.
 * @returns the registry
 */
export function createFolderRegistry (): FolderRegistry {
  const folders = new Map<string, FolderState>()
  const states = (): FolderState[] => Array.from(folders, ([, state]) => state)

  return {
    folders,
    states,
    ready: () => states().filter((state) => state.phase === 'ready' && state.session !== undefined),
    stop:  (root) => {
      const state = folders.get(root)
      state?.session?.stop()
      if (state !== undefined) state.session = undefined
    },
  }
}
