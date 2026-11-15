/** What the person decided to swap: one choice of one project. */
export interface SwapRequest {
  project: string
  /** The question to change, such as `test`. */
  axis:    string
  /** The answer that replaces the project's choice on the axis, such as `vitest`. */
  choice:  string
}

/** What the person decided after reading the plan. */
export type SwapAnswer = 'swap' | 'swap-commit' | 'cancel'

/** How swapping ended: done (`lines` are what smith printed), refused with smith's words and fix, or cancelled by the person. */
export type SwapOutcome =
  | { kind: 'swapped', request: SwapRequest, lines: string[], committed: boolean } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
