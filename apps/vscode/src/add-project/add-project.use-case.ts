import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import { addArguments, followsSiblings } from './add-arguments.mapper.ts'
import type { AddOptions, AddOutcome, AddRequest, PlanAnswer } from './add-options.contract.ts'
import { parseMissingTool } from './add-missing-tool.mapper.ts'
import { parseChoices, parseKinds, parseStacks } from './add-options.mapper.ts'
import { fitOutcomes, parseOutcomes } from './add-outcomes.mapper.ts'

/** What adding a project needs from the editor. */
export interface AddDeps {
  /** Runs smith in the workspace folder for a short answer (`kinds`, `choices`, a dry run). */
  run:     (args: string[]) => Promise<SmithRun>
  /** Runs smith in the workspace folder to its end; `smith add` installs and tests, so it takes long. */
  apply:   (args: string[]) => Promise<SmithRun>
  /**
   * Asks the questions: what is being built, language, stack or choices, name.
   * Resolves undefined when the person cancels, or `compose` when they chose
   * to compose a stack instead (the caller opens the composer).
   */
  ask:     (options: AddOptions) => Promise<AddRequest | 'compose' | undefined>
  /** Shows the plan and asks. `followsSiblings` tells whether `fresh` would change the project. */
  confirm: (request: AddRequest, plan: string, followsSiblings: boolean) => Promise<PlanAnswer>
}

async function readOptions (run: AddDeps['run']): Promise<AddOptions | string> {
  const [kinds, choices, stacks, catalogue] = await Promise.all([run(['kinds', '--json']), run(['choices', '--json']), run(['stacks', '--json']), run(['compose', '--json'])])
  const failed = [kinds, choices, stacks].find((result) => result.code !== 0)
  if (failed !== undefined) return describeFailure(failed)
  try {
    const kindList = parseKinds(kinds.stdout)
    // The outcomes are smith's own list. A smith that cannot print it leaves the long way and the composer.
    const listed = catalogue.code === 0 ? parseOutcomes(catalogue.stdout) : { outcomes: [], languages: [] }

    return { kinds: kindList, choices: parseChoices(choices.stdout), stacks: parseStacks(stacks.stdout), outcomes: fitOutcomes(listed.outcomes, kindList), languages: listed.languages }
  } catch {
    return 'smith printed kinds, choices or stacks in a form this editor does not read. Update the extension, or run smith add in a terminal.'
  }
}

/**
 * Adds a project: reads what smith can add, asks the person, shows the plan of
 * `smith add --dry-run`, and only on a yes runs the real add. smith leaves the
 * files it wrote uncommitted (or commits them when the workspace sets
 * `commit_operations`), so the caller can offer Undo afterwards.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function addProject (deps: AddDeps): Promise<AddOutcome> {
  const options = await readOptions(deps.run)
  if (typeof options === 'string') return { kind: 'refused', message: options }

  const request = await deps.ask(options)
  if (request === 'compose') return { kind: 'composing' }
  if (request === undefined) return { kind: 'cancelled' }

  const plan = await deps.run(addArguments(request, 'plan'))
  if (plan.code !== 0) return { kind: 'refused', message: describeFailure(plan) }

  const answer = await deps.confirm(request, plan.stdout.trim(), followsSiblings(plan.stdout))
  if (answer === 'cancel') return { kind: 'cancelled' }

  const committed = answer.endsWith('-commit')
  const done = await deps.apply(addArguments(request, 'apply', answer.startsWith('fresh') || request.fresh, committed))

  const lines = done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean)
  if (done.code === 0) return { kind: 'added', name: request.name, role: request.role, operation: `add ${request.role} ${request.name}`, lines, committed }
  const tool = parseMissingTool(`${done.stdout}\n${done.stderr}`)

  return tool === undefined ? { kind: 'refused', message: describeFailure(done) } : { kind: 'unchecked', name: request.name, tool, lines }
}
