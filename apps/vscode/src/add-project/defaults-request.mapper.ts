import type { AddRequest } from './add-options.contract.ts'
import { isWebApp, startingEnvironments } from './web-environments.mapper.ts'

/** What the first questions found out, before the person chooses defaults or customising. */
export interface FoundAdd {
  role:       string
  language:   string
  /** The framework, shell or host the person picked in the first question (empty when they decided later). */
  preset:     readonly string[]
  outcomeId?: string
}

/**
 * Whether the person may take the defaults: a function app has a trigger and a
 * runtime that smith does not choose for them, so its questions stay.
 * @param found - what the first questions found out
 */
export function canStartWithDefaults (found: Pick<FoundAdd, 'outcomeId'>): boolean {
  return found.outcomeId !== 'function'
}

/**
 * The request that "Start with the defaults" makes: what the person picked is kept, the
 * shape of the sibling projects is followed, and a web app is built for the workspace's
 * environments (lcl, dev, uat and prd when it has none) with every default of smith. The
 * plan then shows each default as a flag.
 * @param found - what the first questions found out
 * @param name - the name the person typed
 * @param workspaceEnvironments - `[workspace] environments`, or empty
 */
export function defaultsRequest (found: FoundAdd, name: string, workspaceEnvironments: readonly string[]): AddRequest {
  const request: AddRequest = { role: found.role, name, language: found.language, with: [...found.preset], fresh: false }

  return isWebApp(request) ? { ...request, environments: startingEnvironments(workspaceEnvironments) } : request
}
