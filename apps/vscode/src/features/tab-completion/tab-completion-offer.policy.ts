/**
 * What the extension remembers, for the person and not for one workspace, about the offer of tab completion:
 * `never` after "Don't ask again", `done` after "Set up", nothing before.
 */
export type TabCompletionMemory = 'never' | 'done' | undefined

/** What decides whether to offer the setup of tab completion. */
export interface TabCompletionFacts {
  /** What the person answered before. */
  remembered:       TabCompletionMemory
  /** The offer was shown in this session and answered "Not now". */
  askedThisSession: boolean
  /** What `smith completion status --json` said: nothing is installed and the person did not say no on the command line. */
  wanted:           boolean
}

/**
 * Offers tab completion only when smith says it is not set up, the person never answered "Don't ask again" or "Set up",
 * and the offer was not shown in this session.
 * @param facts - what is known
 */
export function shouldOfferTabCompletion (facts: TabCompletionFacts): boolean {
  return facts.wanted && facts.remembered === undefined && !facts.askedThisSession
}
