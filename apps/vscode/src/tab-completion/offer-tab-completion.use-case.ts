import type { SmithRun } from '../smith-run/index.ts'
import { parseCompletionStatus } from './completion-status.mapper.ts'
import { shouldOfferTabCompletion, type TabCompletionMemory } from './tab-completion-offer.policy.ts'

/** What the person answers to the offer. */
export type TabCompletionAnswer = 'setup' | 'later' | 'never'

/** What the offer needs from the editor. */
export interface TabCompletionOfferDeps {
  /** What the person answered before, kept for the person in the editor. */
  recall:           () => TabCompletionMemory
  /** Keeps an answer. */
  remember:         (answer: Exclude<TabCompletionMemory, undefined>) => PromiseLike<void>
  /** Whether the offer was shown in this session. */
  askedThisSession: () => boolean
  /** Notes that the offer was shown in this session. */
  markAsked:        () => void
  /** Runs `smith completion status --json`. */
  status:           () => Promise<SmithRun>
  /** Shows the offer; undefined when the person closed it. */
  ask:              () => Promise<TabCompletionAnswer | undefined>
  /** Runs `smith completion install` in the terminal, where the person answers its question. */
  install:          () => void
}

/**
 * Offers tab completion once. Nothing is written by the extension: "Set up" opens the terminal on
 * `smith completion install`, which shows the files and asks. "Not now" asks again in the next session,
 * "Don't ask again" never.
 * @param deps - the editor
 * @returns true when the offer was shown
 */
export async function offerTabCompletion (deps: TabCompletionOfferDeps): Promise<boolean> {
  if (deps.recall() !== undefined || deps.askedThisSession()) return false
  const run = await deps.status()
  const status = run.code === 0 ? parseCompletionStatus(run.stdout) : undefined
  if (status === undefined || !shouldOfferTabCompletion({ remembered: undefined, askedThisSession: false, wanted: status.wanted })) return false
  deps.markAsked()
  const answer = await deps.ask()
  if (answer === 'setup') {
    await deps.remember('done')
    deps.install()
  } else if (answer === 'never') {
    await deps.remember('never')
  }

  return true
}
