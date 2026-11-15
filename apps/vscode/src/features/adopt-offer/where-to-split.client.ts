import * as vscode from 'vscode'
import { runSmith } from '../agent-install/index.ts'
import { say } from '../words/index.ts'
import { dependencyQuestion, lintQuestion, toSplitAnswer, type SplitAnswer, type SplitQuestion } from './split-choices.policy.ts'
import { parseSplitPlan } from './split-plan.mapper.ts'

const TITLE = say('vscode.adopt-offer.monosmith-bring-in')

async function pick<T extends string> (question: SplitQuestion<T>, step: string): Promise<T | undefined> {
  const picked = await vscode.window.showQuickPick(
    question.choices.map((choice) => ({ label: choice.label, description: choice.description, value: choice.value })),
    { title: `${say('vscode.adopt-offer.split-title', { title: TITLE })} (${step})`, placeHolder: `${question.title} ${say('vscode.adopt-offer.split-first-answer')}` },
  )

  return picked?.value
}

/**
 * Asks what smith cannot place when the project moves into `dir`, the
 * questions `smith adopt --init` asks in a terminal: one quick pick for each
 * development dependency it does not know as a tool (the project or the
 * workspace), and one for the project's own ESLint setup (keep it, or use the
 * shared config). It reads the questions from `smith adopt --dry-run --json
 * --move <dir>`, which changes nothing. The first answer of each changes the
 * least. When smith has nothing to ask, or cannot say, it asks nothing.
 * @param executable - the smith to run
 * @param root - the repository root
 * @param dir - the folder the person chose for the project
 * @returns the answers, or undefined when the person pressed Escape
 */
export async function askWhatToSplit (executable: string, root: string, dir: string): Promise<SplitAnswer | undefined> {
  const run = await runSmith(executable, ['adopt', '--dry-run', '--json', '--move', dir], root)
  const plan = run.code === 0 ? parseSplitPlan(run.stdout) : undefined
  if (plan === undefined || (plan.unsure.length === 0 && !plan.lint)) return { hoist: [], eslint: undefined }
  const total = plan.unsure.length + (plan.lint ? 1 : 0)
  const dependencies: Array<{ name: string, goesTo: 'project' | 'workspace' }> = []
  for (const name of plan.unsure) {
    const goesTo = await pick(dependencyQuestion(name), `${dependencies.length + 1}/${total}`)
    if (goesTo === undefined) return undefined
    dependencies.push({ name, goesTo })
  }
  if (!plan.lint) return toSplitAnswer(dependencies, undefined)
  const lint = await pick(lintQuestion(), `${total}/${total}`)

  return lint === undefined ? undefined : toSplitAnswer(dependencies, lint)
}
