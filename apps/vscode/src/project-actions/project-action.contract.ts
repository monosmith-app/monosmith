/** How an action on a project ended: done (`lines` are what smith printed), refused with smith's words, or cancelled by the person. */
export type ProjectActionOutcome =
  | { kind: 'done', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }

/** The kind of a project: what `smith kind` can make it. */
export type ProjectKind = 'app' | 'lib' | 'package'
