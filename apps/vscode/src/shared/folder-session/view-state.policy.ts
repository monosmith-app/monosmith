import type { FolderPhase, ViewState } from './folder-state.contract.ts'

/** What the view shows when no folder is ready, most pressing first. */
const PRIORITY: FolderPhase[] = ['starting', 'untrusted', 'missing', 'failed']

/** How many open folders have no `smith.toml`, by what they hold. */
export interface FoldersWithoutSmith {
  /** Repositories smith can convert (Nx, Turborepo, other monorepo tools). */
  adoptable?:    number
  /** Folders with nothing in them. */
  empty?:        number
  /** Folders with files but no tool smith recognises. */
  unrecognised?: number
}

/**
 * What the view shows for the phases of all its folders. One ready folder is
 * enough to show projects; otherwise the welcome content of the most pressing
 * state: still starting, then not trusted, then smith missing, then failed.
 * With no folder that has a `smith.toml`: `adoptable` when one looks like a
 * repository smith can bring in, else `empty` for an empty folder, else
 * `unrecognised` for one with files, and `nofolder` when nothing is open.
 */
export function judgeViewState (phases: readonly FolderPhase[], without: FoldersWithoutSmith = {}): ViewState {
  if (phases.length === 0) {
    if ((without.adoptable ?? 0) > 0) return 'adoptable'
    if ((without.empty ?? 0) > 0) return 'empty'

    return (without.unrecognised ?? 0) > 0 ? 'unrecognised' : 'nofolder'
  }

  return phases.includes('ready') ? 'ready' : PRIORITY.find((phase) => phases.includes(phase)) ?? 'failed'
}
