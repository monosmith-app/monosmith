import { appliesWhen, askableQuestions, parseQuestions, type PickAnswer } from '../smith-questions/index.ts'
import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import { toConvertArguments } from './convert-arguments.mapper.ts'
import type { ConvertOutcome } from './convert-workspace.contract.ts'
import { say } from '../words/index.ts'

/** What converting a workspace needs from the editor. */
export interface ConvertDeps {
  /** Runs smith in the folder for a read (`convert --questions --json`, `convert … --dry-run`). */
  run:       (args: string[]) => Promise<SmithRun>
  /** Runs smith in the folder to its end (`convert … --yes`). */
  apply:     (args: string[]) => Promise<SmithRun>
  /** Shows one question and returns the answer, or undefined when the person cancels. */
  pick:      PickAnswer
  /** Asks the unticked "Commit this change so it can be undone" box. Resolves undefined when the person cancels. */
  askCommit: () => Promise<boolean | undefined>
  /** Shows the plan and asks. Nothing was changed yet. */
  confirm:   (plan: string) => Promise<boolean>
}

/**
 * Converts an Nx or Turborepo workspace: asks the questions of `smith convert --questions --json` in
 * smith's order (a question only when its `when` holds), shows the plan of `--dry-run`, asks the commit
 * box, and only after a yes runs `--yes`. smith commits only when the box is ticked.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function convertWorkspace (deps: ConvertDeps): Promise<ConvertOutcome> {
  const listed = await deps.run(['convert', '--questions', '--json'])
  if (listed.code !== 0) return { kind: 'refused', message: say('vscode.convert-workspace.update-smith-and-try-again', { describeFailure: describeFailure(listed) }) }
  let questions
  try {
    questions = askableQuestions(parseQuestions(listed.stdout))
  } catch {
    return { kind: 'refused', message: say('vscode.convert-workspace.smith-did-not-list-the-questions') }
  }
  const answers: Record<string, string> = {}
  for (const question of questions) {
    if (!appliesWhen(question, answers)) continue
    const picked = await deps.pick(question, [answers[question.key] ?? question.default])
    if (picked === undefined) return { kind: 'cancelled' }
    answers[question.key] = picked.join(',')
  }

  const plan = await deps.run(toConvertArguments(answers, false))
  if (plan.code !== 0) return { kind: 'refused', message: describeFailure(plan) }
  const commit = await deps.askCommit()
  if ((commit === undefined) || !await deps.confirm(plan.stdout.trim())) return { kind: 'cancelled' }

  const done = await deps.apply(toConvertArguments(answers, true, commit))
  const lines = done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean)

  return done.code === 0 ? { kind: 'converted', lines, committed: commit } : { kind: 'refused', message: describeFailure(done) }
}
