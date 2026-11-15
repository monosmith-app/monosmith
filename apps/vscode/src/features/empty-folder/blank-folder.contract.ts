/**
 * What a folder without a `smith.toml` and without a tool smith recognises
 * holds: `empty` (nothing a person put there) or `unrecognised` (files, but
 * no monorepo tool smith can convert).
 */
export type BlankKind = 'empty' | 'unrecognised'

/** A folder MonoSmith can offer to start in (`empty`) or to bring in (`unrecognised`). */
export interface BlankFolder {
  /** Its path on disk. */
  root: string
  name: string
  kind: BlankKind
}
