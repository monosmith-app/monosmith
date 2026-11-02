import { describeFailure, parseDoctorFailure, type SmithRun } from '../smith-run/index.ts'
import { composeArguments, describePlan, parseReport } from './compose-report.mapper.ts'
import type { ComposeAnswer, ComposeItem, ComposeOption, ComposeOutcome, ComposeQuestion, ComposeReport } from './compose-stack.contract.ts'

/** What composing a stack needs from the editor. */
export interface ComposeDeps {
  /** Runs smith in the workspace folder for a short answer (`compose --json`). */
  run:     (args: string[]) => Promise<SmithRun>
  /** Runs smith to its end (`add`, `recipe new`); the add installs and tests. */
  apply:   (args: string[]) => Promise<SmithRun>
  /**
   * Shows the catalogue with the picks ticked and the last message (a conflict,
   * what is missing). Resolves the ids ticked, or undefined when the person cancels.
   */
  pick:    (items: ComposeItem[], ticked: readonly string[], message: string) => Promise<string[] | undefined>
  /** Asks the open question; the suggested answer is pre-selected, never taken. Resolves the answer or undefined. */
  answer:  (question: ComposeQuestion) => Promise<ComposeOption | undefined>
  /** Asks the project name. Resolves undefined when the person cancels. */
  askName: () => Promise<string | undefined>
  /** Tells a message that does not stop the flow (a plugin that cannot be ticked). */
  tell:    (message: string) => void
  /** Shows the plan and asks what to do with it. */
  confirm: (planText: string, hasRecipe: boolean) => Promise<ComposeAnswer>
}

async function read (deps: ComposeDeps, args: string[]): Promise<ComposeReport | string> {
  const result = await deps.run(args)
  if (result.code !== 0) return describeFailure(result)
  try {
    return parseReport(result.stdout)
  } catch {
    return 'smith printed its catalogue in a form this editor does not read. Update the extension, or run smith compose in a terminal.'
  }
}

/** The ticked items that go into --with, and the language the ticks fix. */
function split (items: readonly ComposeItem[], ids: readonly string[]): { picks: string[], language: string, plugins: ComposeItem[] } {
  const ticked = ids.map((id) => items.find((item) => item.id === id)).filter((item): item is ComposeItem => item !== undefined)
  const language = ticked.find((item) => item.kind === 'language')?.name ?? ticked.find((item) => item.kind === 'choice')?.language ?? ''

  return {
    picks:   ticked.filter((item) => item.kind === 'choice').map((item) => item.name),
    language,
    plugins: ticked.filter((item) => item.kind === 'plugin'),
  }
}

/**
 * Composes a stack: shows the catalogue, checks the ticks with smith after
 * each round (a conflict is explained, an open question is asked, nothing is
 * chosen silently), then names the project and shows the plan. The plan can be
 * run, copied or kept as a recipe. Nothing is written before the person says so.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function composeStack (deps: ComposeDeps): Promise<ComposeOutcome> {
  const first = await read(deps, composeArguments({ picks: [] }))
  if (typeof first === 'string') return { kind: 'refused', message: first }
  let ticked: string[] = []
  let message = ''

  for (;;) {
    const chosen = await deps.pick(first.items, ticked, message)
    if (chosen === undefined || chosen.length === 0) return { kind: 'cancelled' }
    ticked = chosen
    const { picks, language, plugins } = split(first.items, ticked)
    for (const plugin of plugins) deps.tell(`${plugin.title} is not part of one project. Turn it on with: ${plugin.route}`)
    ticked = ticked.filter((id) => plugins.every((plugin) => plugin.id !== id))
    if (picks.length === 0) {
      message = 'Tick the choice that writes the project too: a framework, a desktop shell or a cloud host.'
      continue
    }

    const checked = await read(deps, composeArguments({ language, picks }))
    if (typeof checked === 'string') return { kind: 'refused', message: checked }
    const check = checked.check
    if (check === undefined) return { kind: 'refused', message: 'smith did not check the selection. Update smith, or run smith compose in a terminal.' }
    if (check.conflict !== '') {
      message = `Conflict: ${check.conflict}`
      continue
    }
    if (check.question !== undefined) {
      const option = await deps.answer(check.question)
      const asked = option === undefined ? undefined : first.items.find((item) => item.kind === 'choice' && item.name === option.name && item.language === check.language)
      if (asked !== undefined) ticked = [...ticked, asked.id]
      message = asked === undefined ? `${check.question.choice} still needs a choice on the ${check.question.axis} axis.` : ''
      continue
    }
    if (!check.complete) {
      message = check.missing
      continue
    }

    const name = await deps.askName()
    if (name === undefined) return { kind: 'cancelled' }
    const planned = await read(deps, composeArguments({ language: check.language, picks: check.selection, kind: 'app', name }))
    if (typeof planned === 'string') return { kind: 'refused', message: planned }
    const plan = planned.plan
    if (plan === undefined || plan.argv.length < 2) return { kind: 'refused', message: 'smith made no plan for these picks.' }

    const answer = await deps.confirm(describePlan(plan), plan.recipeArgv.length > 1)
    switch (answer) {
      case 'cancel': {
        return { kind: 'cancelled' }
      }
      case 'copy': {
        return { kind: 'copied', command: plan.command }
      }
      case 'recipe': {
        const done = await deps.apply(plan.recipeArgv.slice(1))

        return done.code === 0 ? { kind: 'recipe', lines: lines(done.stdout) } : { kind: 'refused', message: describeFailure(done) }
      }
      default: {
        const committed = answer === 'run-commit'
        const done = await deps.apply([...plan.argv.slice(1), ...(committed ? ['--commit'] : []), '--yes'])
        if (done.code === 0) return { kind: 'added', name, lines: lines(done.stdout), committed }

        // The project is on disk and fails its checks: not a refusal, and not a cancel.
        const everything = `${done.stdout}\n${done.stderr}`
        const failure = parseDoctorFailure(everything)

        return failure === undefined ? { kind: 'refused', message: describeFailure(done) } : { kind: 'written', name, operation: `add app ${name}`, failure, lines: lines(everything) }
      }
    }
  }
}

function lines (stdout: string): string[] {
  return stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean)
}
