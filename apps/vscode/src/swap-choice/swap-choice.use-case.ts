import type { ChoiceOption } from '../add-project/index.ts'
import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import { swapArguments } from './swap-arguments.mapper.ts'
import type { SwapAnswer, SwapOutcome, SwapRequest } from './swap-choice.contract.ts'

/** What swapping a choice needs from the editor. */
export interface SwapDeps {
  /** Runs smith in the workspace folder for a short answer (`choices`, a dry run). */
  run:     (args: string[]) => Promise<SmithRun>
  /** Runs smith in the workspace folder to its end; the swap installs and tests, so it takes long. */
  apply:   (args: string[]) => Promise<SmithRun>
  /** Reads `smith choices --json` (the Add flow's parser). */
  parse:   (json: string) => ChoiceOption[]
  /** Asks which project, which axis and which choice. Resolves undefined when the person cancels. */
  ask:     (choices: ChoiceOption[]) => Promise<SwapRequest | undefined>
  /** Shows the plan and asks. */
  confirm: (request: SwapRequest, plan: string) => Promise<SwapAnswer>
}

/**
 * Swaps one choice of a project, such as its test runner. It reads the choices
 * smith offers, asks the person, shows the plan of `smith swap --dry-run`, and
 * only on a yes runs the swap. smith leaves the files uncommitted unless the
 * person asked for the commit, so the caller can offer Undo afterwards.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function swapChoice (deps: SwapDeps): Promise<SwapOutcome> {
  const listed = await deps.run(['choices', '--json'])
  if (listed.code !== 0) return { kind: 'refused', message: describeFailure(listed) }
  let choices: ChoiceOption[]
  try {
    choices = deps.parse(listed.stdout)
  } catch {
    return { kind: 'refused', message: 'smith printed its choices in a form this editor does not read. Update the extension, or run smith swap in a terminal.' }
  }

  const request = await deps.ask(choices)
  if (request === undefined) return { kind: 'cancelled' }

  const plan = await deps.run(swapArguments(request, 'plan'))
  if (plan.code !== 0) return { kind: 'refused', message: describeFailure(plan) }

  const answer = await deps.confirm(request, plan.stdout.trim())
  if (answer === 'cancel') return { kind: 'cancelled' }

  const committed = answer === 'swap-commit'
  const done = await deps.apply(swapArguments(request, 'apply', committed))

  return done.code === 0
    ? { kind: 'swapped', request, lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean), committed }
    : { kind: 'refused', message: describeFailure(done) }
}
