import type { FolderPhase, ViewState } from './folder-state.contract.ts'

/** What to show when no folder is ready, most pressing first. */
const PRIORITY: FolderPhase[] = ['starting', 'untrusted', 'missing', 'failed']

/**
 * What the view shows for the phases of all its folders. One ready folder is
 * enough to show projects; otherwise the welcome content of the most pressing
 * state: still starting, then not trusted, then smith missing, then failed.
 * No folder with a `smith.toml` at all is `empty`.
 */
export function judgeViewState (phases: readonly FolderPhase[]): ViewState {
  if (phases.length === 0) return 'empty'

  return phases.includes('ready') ? 'ready' : PRIORITY.find((phase) => phases.includes(phase)) ?? 'failed'
}
