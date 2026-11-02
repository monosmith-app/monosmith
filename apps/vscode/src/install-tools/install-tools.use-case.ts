import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import type { ToolOption, ToolPlan } from './tool-plan.contract.ts'

/** The answer to the administrator question. */
export type AdminAnswer = 'myself' | 'run' | 'skip'

/** What installing the tools needs from the editor. */
export interface InstallToolsDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:           (args: string[]) => Promise<SmithRun>
  /** Asks which way to get the tool; the options show their command. Returns the option's `value`, or undefined when the person cancels. */
  chooseWay:     (plan: ToolPlan) => Promise<string | undefined>
  /** Asks what to do with a command that needs administrator rights. Returns undefined when the person cancels. */
  chooseAdmin:   (plan: ToolPlan, option: ToolOption) => Promise<AdminAnswer | undefined>
  /** Opens the official page. */
  openPage:      (url: string) => Promise<void>
  /** Shows text the person copies: the commands that would work. */
  showCommands:  (plan: ToolPlan, lines: string[]) => Promise<void>
  /** Runs smith in a terminal, where sudo or the system can ask for a password. */
  runInTerminal: (args: string[]) => void
}

/** What happened to one tool. */
export interface ToolResult {
  tool:   string
  /** The `value` of the option taken. */
  choice: string
  /** What to tell the person: smith's own lines, or what to do next. */
  lines:  string[]
}

/** How installing ended: nothing was missing, smith refused with its words, or the tools were handled one by one. */
export type InstallToolsOutcome =
  | { kind: 'nothing' } |
  { kind: 'refused', message: string } |
  { kind: 'handled', results: ToolResult[] }

/**
 * Offers the ways to get each tool that is missing, and carries out the answers.
 * smith prints the plan (`tools install --dry-run --json`), the person picks
 * for each tool with the command in front of them, and smith runs only that
 * command. A command that needs administrator rights is never run before a
 * second question; the "run it" answer opens a terminal so the system or sudo can ask.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function installTools (deps: InstallToolsDeps): Promise<InstallToolsOutcome> {
  const planned = await deps.run(['tools', 'install', '--dry-run', '--json'])
  if (planned.code !== 0) return { kind: 'refused', message: describeFailure(planned) }
  let plans: ToolPlan[]
  try {
    plans = (JSON.parse(planned.stdout) as ToolPlan[] | null) ?? []
  } catch {
    return { kind: 'refused', message: 'smith tools install printed something that is not a plan' }
  }
  if (plans.length === 0) return { kind: 'nothing' }
  const results: ToolResult[] = []
  for (const plan of plans) {
    const value = await deps.chooseWay(plan)
    if (value === undefined) break
    results.push(await carryOut(deps, plan, value))
  }

  return { kind: 'handled', results }
}

async function carryOut (deps: InstallToolsDeps, plan: ToolPlan, value: string): Promise<ToolResult> {
  const option = plan.options.find((o) => o.value === value)
  const done = (lines: string[]): ToolResult => ({ tool: plan.name, choice: value, lines })
  if (option === undefined) return done([])
  switch (option.kind) {
    case 'page': {
      await deps.openPage(plan.download)

      return done([`Opened ${plan.download}. Install ${plan.title}, then open a new terminal.`])
    }
    case 'myself': {
      await deps.showCommands(plan, commandsOf(plan))

      return done([`Install ${plan.title} yourself, then open a new terminal.`])
    }
    case 'skip': {
      return done([`Skipped ${plan.title}: ${plan.consequence}.`])
    }
    default: {
      return await install(deps, plan, option)
    }
  }
}

/** Every install command that would work, then the official page, one per line. */
function commandsOf (plan: ToolPlan): string[] {
  const commands = plan.options.filter((o) => o.kind === 'install' && o.command !== undefined).map((o) => (o.command ?? []).join(' '))

  return [...commands, plan.download]
}

async function install (deps: InstallToolsDeps, plan: ToolPlan, option: ToolOption): Promise<ToolResult> {
  const command = (option.command ?? []).join(' ')
  const args = ['tools', 'install', plan.name, '--manager', option.manager ?? '']
  const done = (lines: string[]): ToolResult => ({ tool: plan.name, choice: option.value, lines })
  if (option.admin === true) {
    const answer = await deps.chooseAdmin(plan, option)
    if (answer === 'run') {
      deps.runInTerminal([...args, '--with-admin'])

      return done([`Running in the terminal: ${command}`])
    }

    return done([answer === 'myself' ? `Run this yourself, then open a new terminal: ${command}` : `Skipped ${plan.title}: ${plan.consequence}.`])
  }
  const run = await deps.run(args)

  return done(run.code === 0 ? run.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) : [`The install did not finish. ${describeFailure(run)}`])
}
