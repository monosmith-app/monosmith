/** What a folder without a `smith.toml` looks like. `repository` is any other repository smith can adopt. */
export type RepositoryKind = 'nx' | 'turborepo' | 'repository'

/** A folder MonoSmith can offer to bring in. */
export interface RepositoryOffer {
  /** Its path on disk. */
  root:   string
  name:   string
  kind:   RepositoryKind
  /** The file that gave it away (`nx.json`), for the log. */
  marker: string
}
