/** What the person asked to add to a project. */
export interface InstallRequest {
  /** The project that gains the dependency. */
  project:  string
  /** A library of this workspace, or a package from the registry. */
  source:   'workspace' | 'registry'
  /** The library, when `source` is `workspace`. */
  library?: string
  /** The version range written for the library; absent lets smith follow the workspace. */
  range?:   string
  /** The registry packages, when `source` is `registry`. */
  packages: string[]
  /** A development dependency (`-D`), for a registry package. */
  dev:      boolean
}

/** A project, as far as the rules of a link read it. */
export interface LinkNode {
  name:     string
  kind:     string
  language: string
  deps:     string[]
}

/** How adding a dependency ended. */
export type InstallOutcome =
  { kind: 'added', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
