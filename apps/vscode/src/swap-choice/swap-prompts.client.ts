import * as vscode from 'vscode'
import { choicesByAxis, type ChoiceOption } from '../add-project/index.ts'
import type { SwapAnswer, SwapOutcome, SwapRequest } from './swap-choice.contract.ts'

const TITLE = 'MonoSmith: Swap a choice'

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
    ? await vscode.window.showQuickPick(projects.map((item) => ({ label: item.name, description: item.language })), { title: `${TITLE} (1/3): which project?`, placeHolder: 'The project whose decision changes' })
    : { label: project, description: projects.find((item) => item.name === project)?.language ?? '' }
  if (target === undefined) return undefined

  const axes = choicesByAxis(choices, target.description).filter((axis) => axis.choices.length > 1)
  if (axes.length === 0) {
    void vscode.window.showInformationMessage(`MonoSmith: ${target.label} has no question with more than one answer. smith choices lists what can change.`)

    return undefined
  }
  const axis = await vscode.window.showQuickPick(axes.map((item) => ({ label: item.axis, description: item.choices.map((choice) => choice.name).join(', ') })), { title: `${TITLE} (2/3): which decision?`, placeHolder: 'The test runner, the linter, the style' })
  if (axis === undefined) return undefined

  const chosen = axes.find((item) => item.axis === axis.label)?.choices ?? []
  const choice = await vscode.window.showQuickPick(chosen.map((item) => ({ label: item.name, description: item.description })), { title: `${TITLE} (3/3): replace it with?`, placeHolder: 'The answer the project moves to' })

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
  const swap = 'Swap'
  const commit = 'Swap and commit (so Undo works)'
  const open = 'Open the plan'
  for (;;) {
    const choice = await vscode.window.showInformationMessage(
      `MonoSmith: move ${request.project} to ${request.choice} (${request.axis})?`,
      { modal: true, detail: `${plan}\n\nNothing was changed yet. smith removes what the old choice wrote, writes the new one, installs the packages and runs the project's build and tests. By default smith leaves the files uncommitted: review them with git. The commit button makes it one commit that Undo can revert.` },
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

/**
 * Says how swapping ended, and offers Undo after a committed success. It does not wait for the person to read it.
 * @param outcome - what `swapChoice` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellSwapOutcome (outcome: SwapOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'swapped': {
      for (const line of outcome.lines) onOutput(line)
      const undo = 'Undo'
      const left = outcome.lines.filter((line) => line.startsWith('Left to do by hand'))
      void vscode.window.showInformationMessage(
        `MonoSmith: ${outcome.request.project} now uses ${outcome.request.choice}. ${left.length > 0 ? `${left.length} thing${left.length === 1 ? ' is' : 's are'} left to do by hand: see the output. ` : ''}${outcome.committed ? 'Committed as one change: Undo reverts it.' : 'Not committed: review with git status.'}`,
        ...(outcome.committed ? [undo] : []),
      ).then(async (choice) => {
        if (choice === undo) await vscode.commands.executeCommand('monosmith.undo')
      })
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was swapped. ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
