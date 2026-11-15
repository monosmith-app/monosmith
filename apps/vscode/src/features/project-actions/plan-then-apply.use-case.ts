import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import type { ProjectActionOutcome } from './project-action.contract.ts'

/** What an action with a plan needs from the editor. */
export interface PlanThenApplyDeps<Answer> {
  /** Runs smith for a short answer (the plan). */
  run:     (args: string[]) => Promise<SmithRun>
  /** Runs smith to its end (the change), with progress shown. */
  apply:   (args: string[]) => Promise<SmithRun>
  /** Shows the plan and asks; undefined means the person cancelled. */
  confirm: (plan: string) => Promise<Answer | undefined>
}

/**
 * Shows what smith would do (the plan, or a fixed text when smith has no plan command),
 * asks, and only on a yes runs the change with `--yes`. smith never commits here:
 * the caller offers the Commit command afterwards.
 * @param deps - how to run smith and ask the person
 * @param planArguments - the arguments of the plan (`--dry-run`), or undefined when smith has none
 * @param applyArguments - the arguments of the change for the answer the person gave
 * @returns how it ended; it never throws
 */
export async function planThenApply<Answer> (deps: PlanThenApplyDeps<Answer>, planArguments: string[] | undefined, applyArguments: (answer: Answer) => string[]): Promise<ProjectActionOutcome> {
  let plan = ''
  if (planArguments !== undefined) {
    const planned = await deps.run(planArguments)
    if (planned.code !== 0) return { kind: 'refused', message: describeFailure(planned) }
    plan = planned.stdout.trim()
  }
  const answer = await deps.confirm(plan)
  if (answer === undefined) return { kind: 'cancelled' }

  const done = await deps.apply(applyArguments(answer))

  return done.code === 0 ? { kind: 'done', lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) } : { kind: 'refused', message: describeFailure(done) }
}
