import * as vscode from 'vscode'
import { choicesByAxis, type ChoiceOption } from '../add-project/index.ts'
import type { SwapAnswer, SwapOutcome, SwapRequest } from './swap-choice.contract.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.swap-choice.monosmith-swap-a-choice')

/**
 * Asks what to swap with quick picks: the project (unless it is given), the
 * axis it has a choice on, and the choice that replaces the current one.
 * @param choices - what `smith choices --json` lists
 * @param projects - the workspace's projects
 * @param project - the project the command was started on, when the person used a project's menu
 * @returns the request, or undefined when the person pressed Escape at any question
 */
export async function askSwapRequest (choices: readonly ChoiceOption[], projects: ReadonlyArray<{ name: string, language: string }>, project?: string): Promise<SwapRequest | undefined> {
  const target = project === undefined
    ? await vscode.window.showQuickPick(projects.map((item) => ({ label: item.name, description: item.language })), { title: say('vscode.install-dependency.1-3-which-project', { title: TITLE }), placeHolder: say('vscode.swap-choice.the-project-whose-decision-changes') })
    : { label: project, description: projects.find((item) => item.name === project)?.language ?? '' }
  if (target === undefined) return undefined

  const axes = choicesByAxis(choices, target.description).filter((axis) => axis.choices.length > 1)
  if (axes.length === 0) {
    void vscode.window.showInformationMessage(say('vscode.swap-choice.monosmith-has-no-question-with-more', { label: target.label }))

    return undefined
  }
  const axis = await vscode.window.showQuickPick(axes.map((item) => ({ label: item.axis, description: item.choices.map((choice) => choice.name).join(', ') })), { title: say('vscode.swap-choice.2-3-which-decision', { title: TITLE }), placeHolder: say('vscode.swap-choice.the-test-runner-the-linter-the') })
  if (axis === undefined) return undefined

  const chosen = axes.find((item) => item.axis === axis.label)?.choices ?? []
  const choice = await vscode.window.showQuickPick(chosen.map((item) => ({ label: item.name, description: item.description })), { title: say('vscode.swap-choice.3-3-replace-it-with', { title: TITLE }), placeHolder: say('vscode.swap-choice.the-answer-the-project-moves-to') })

  return choice === undefined ? undefined : { project: target.label, axis: axis.label, choice: choice.label }
}

/**
 * Shows the plan of `smith swap --dry-run` and asks. The commit button ticks
 * the box that is unticked by default: make the swap one commit, so Undo works.
 * @param request - what the person asked for
 * @param plan - the text of the dry run
 * @returns `swap`, `swap-commit` or `cancel`
 */
export async function confirmSwapPlan (request: SwapRequest, plan: string): Promise<SwapAnswer> {
  const swap = say('vscode.swap-choice.swap')
  const commit = say('vscode.swap-choice.swap-and-commit-so-undo-works')
  const open = say('vscode.add-project.open-the-plan')
  for (;;) {
    const choice = await vscode.window.showInformationMessage(
      say('vscode.swap-choice.monosmith-move-to', { project: request.project, choice: request.choice, axis: request.axis }),
      { modal: true, detail: say('vscode.swap-choice.nothing-was-changed-yet-smith-removes', { plan }) },
      swap,
      commit,
      open,
    )
    if (choice === open) {
      const document = await vscode.workspace.openTextDocument({ language: 'plaintext', content: `${plan}\n` })
      await vscode.window.showTextDocument(document, { preview: true })
      continue
    }
    if (choice === swap) return 'swap'

    return choice === commit ? 'swap-commit' : 'cancel'
  }
}

/** The sentence that counts what smith leaves to do by hand: nothing, one thing or several. */
function leftToDo (count: number): string {
  if (count === 0) return ''

  return say('vscode.swap-choice.thing-left-to-do-by-hand', { length: count, part: say(count === 1 ? 'vscode.swap-choice.is' : 'vscode.swap-choice.s-are') })
}

/**
 * Says how swapping ended, and offers Undo after a committed success. It does not wait for the person to read it.
 * @param outcome - what `swapChoice` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellSwapOutcome (outcome: SwapOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'swapped': {
      for (const line of outcome.lines) onOutput(line)
      const undo = say('vscode.undo-change.undo')
      const left = outcome.lines.filter((line) => line.startsWith(say('vscode.swap-choice.left-to-do-by-hand')))
      void vscode.window.showInformationMessage(
        say('vscode.swap-choice.monosmith-now-uses', { project: outcome.request.project, choice: outcome.request.choice, part: leftToDo(left.length), part2: say(outcome.committed ? 'vscode.add-project.committed-as-one-change-undo-reverts' : 'vscode.compose-stack.not-committed-review-with-git-status') }),
        ...(outcome.committed ? [undo] : []),
      ).then(async (choice) => {
        if (choice === undo) await vscode.commands.executeCommand('monosmith.undo')
      })
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.swap-choice.monosmith-nothing-was-swapped', { message: outcome.message }))
      break
    }
    case 'cancelled': {
      break
    }
  }
}
