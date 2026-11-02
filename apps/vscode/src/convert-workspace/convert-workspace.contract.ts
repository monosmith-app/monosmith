/** How converting a workspace ended. */
export type ConvertOutcome =
  | { kind: 'converted', lines: string[], committed: boolean } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
