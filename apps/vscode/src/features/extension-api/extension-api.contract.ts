/**
 * What `activate` hands back: a read-only view of the extension's state, for
 * the integration tests that run in a VS Code host. It is not a public API and
 * may change with any release.
 */
export interface ExtensionStateApi {
  /** What the view shows: `ready`, `starting`, `missing`, `untrusted`, `failed`, `adoptable`, `empty`, `unrecognised` or `nofolder`. */
  viewState:  () => string
  /** The names of the projects of every folder that runs smith. */
  projects:   () => string[]
  /** How the last run left a task id (`demo:build`) in any folder. */
  taskStatus: (task: string) => string | undefined
  /** Whether a run is in progress. */
  isRunning:  () => boolean
  /** The folder roots that run smith. */
  roots:      () => string[]
  /** What the doctor said about the agent files of a folder: `missing`, `current` or `unknown`. */
  agentFiles: (root: string) => string | undefined
  /** The kind of each folder offered to be brought into MonoSmith (`nx`, `turborepo`, `repository`). */
  offers:     () => string[]
}
