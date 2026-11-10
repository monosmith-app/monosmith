/** What smith could not place when the project at the root moves: the questions to ask. */
export interface SplitPlan {
  /** Development dependencies smith does not know as tools. */
  unsure: readonly string[]
  /** True when the project has its own ESLint setup. */
  lint:   boolean
}

/**
 * Reads the questions of the split from the report of
 * `smith adopt --dry-run --json --move <folder>`: the `unsure` and `lint`
 * fields of its `movePlan`. Anything else in the report is ignored.
 * @param stdout - what smith printed
 * @returns the plan, or undefined when smith printed no move plan (or no JSON)
 */
export function parseSplitPlan (stdout: string): SplitPlan | undefined {
  let report: unknown
  try {
    report = JSON.parse(stdout)
  } catch {
    return undefined
  }
  const plan = (report as { movePlan?: { unsure?: unknown, lint?: unknown } } | null)?.movePlan
  if (plan === undefined || plan === null || typeof plan !== 'object') return undefined
  const unsure = Array.isArray(plan.unsure) ? plan.unsure.filter((name): name is string => typeof name === 'string') : []

  return { unsure, lint: plan.lint === true }
}
