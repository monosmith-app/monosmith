import type { RunPlan } from './run-plan.contract.ts'

/** A run that was started, and the folder it was started in. */
export interface LastRun {
  root: string
  plan: RunPlan
}

/** Remembers the last run asked for, so it can be run again. */
export interface LastRunStore {
  remember: (run: LastRun) => void
  last:     () => LastRun | undefined
}

/** A store that remembers nothing yet. */
export function createLastRunStore (): LastRunStore {
  let latest: LastRun | undefined

  return {
    remember: (run) => { latest = run },
    last:     () => latest,
  }
}
