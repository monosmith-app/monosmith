/** What a folder without a `smith.toml` looks like. `repository` is any other repository smith can adopt. */
export type RepositoryKind = 'nx' | 'turborepo' | 'repository'

/** The one project a plain repository keeps at its root, where smith cannot manage it: it has to move into a folder of its own. */
export interface RootProject {
  /** The file that shows it: `package.json`, `go.mod`, `pyproject.toml` or `Cargo.toml`. */
  manifest:  string
  ecosystem: 'node' | 'go' | 'python' | 'rust'
  /** The folder name offered: the manifest name in kebab-case. */
  name:      string
  /** True when it looks like a program, which makes an app folder the suggestion. */
  app:       boolean
}

/** A folder MonoSmith can offer to bring in. */
export interface RepositoryOffer {
  /** Its path on disk. */
  root:     string
  name:     string
  kind:     RepositoryKind
  /** The file that gave it away (`nx.json`), for the log. */
  marker:   string
  /** Set when the folder is one project at its root: bringing it in asks where to move it. */
  project?: RootProject
}
