import * as vscode from 'vscode'
import { checkProjectName } from '../add-project/index.ts'
import { pickLabels } from './compose-report.mapper.ts'
import type { ComposeAnswer, ComposeItem, ComposeOption, ComposeOutcome, ComposeQuestion } from './compose-stack.contract.ts'

const TITLE = 'MonoSmith: Compose a stack'

/**
 * Shows the catalogue as one list, grouped by layer, with the picks ticked.
 * Typing filters by name, description and the words that find an item
 * ("desktop", "serverless").
 * @param items - the catalogue
 * @param ticked - the ids ticked so far
 * @param message - the last conflict or what is missing, shown as the placeholder
 * @returns the ids ticked, or undefined when the person pressed Escape
 */
export async function pickItems (items: readonly ComposeItem[], ticked: readonly string[], message: string): Promise<string[] | undefined> {
  const entries: Array<vscode.QuickPickItem & { id?: string }> = []
  let layer = ''
  for (const item of items) {
    if (item.layer !== layer) {
      layer = item.layer
      entries.push({ label: layer, kind: vscode.QuickPickItemKind.Separator })
    }
    entries.push({ ...pickLabels(item), id: item.id, picked: ticked.includes(item.id) })
  }
  const chosen = await vscode.window.showQuickPick(entries, {
    title:              message === '' ? `${TITLE} (1/3): what do you want?` : `${TITLE}: ${message}`,
    placeHolder:        'Search, then tick what you want: a language, a framework, a desktop shell, a cloud host',
    canPickMany:        true,
    matchOnDescription: true,
    matchOnDetail:      true,
  })

  return chosen?.flatMap((entry) => entry.id === undefined ? [] : [entry.id])
}

/**
 * Asks the question a pick leaves open. The suggested answer comes first and is
 * only marked: the person confirms it.
 * @param question - the axis and its options
 * @returns the answer, or undefined when the person pressed Escape
 */
export async function answerQuestion (question: ComposeQuestion): Promise<ComposeOption | undefined> {
  const options = [...question.options].sort((a, b) => Number(b.suggested) - Number(a.suggested))
  const picked = await vscode.window.showQuickPick(
    options.map((option) => ({ label: `${option.title}${option.suggested ? ' (suggested)' : ''}`, detail: option.description, option })),
    { title: `${TITLE} (2/3): which ${question.axis} for ${question.choice}?`, placeHolder: 'smith does not choose this for you', matchOnDetail: true },
  )

  return picked?.option
}

/**
 * Asks the project name.
 * @param existing - names the workspace has already
 * @returns the name, or undefined when the person pressed Escape
 */
export async function askProjectName (existing: readonly string[]): Promise<string | undefined> {
  return await vscode.window.showInputBox({
    title:         `${TITLE} (3/3): name`,
    prompt:        'The name of the new app',
    placeHolder:   'for example desk',
    validateInput: (value) => checkProjectName(value) ?? (existing.includes(value) ? `${value} exists already: pick another name.` : undefined),
  })
}

/**
 * Shows the plan and asks what to do with it. Nothing was changed yet.
 * @param plan - the words of the plan
 * @param hasRecipe - whether the plugins can be saved as a recipe
 * @returns the answer
 */
export async function confirmComposePlan (plan: string, hasRecipe: boolean): Promise<ComposeAnswer> {
  const run = 'Run it'
  const commit = 'Run it and commit'
  const copy = 'Copy the command'
  const recipe = 'Save the plugins as a recipe'
  const choice = await vscode.window.showInformationMessage(
    'MonoSmith: this is the plan for your stack.',
    { modal: true, detail: `${plan}\n\nNothing was changed yet. By default smith leaves the files uncommitted: review them with git.` },
    run,
    commit,
    copy,
    ...(hasRecipe ? [recipe] : []),
  )
  switch (choice) {
    case run: { return 'run'
    }
    case commit: { return 'run-commit'
    }
    case copy: { return 'copy'
    }
    case recipe: { return 'recipe'
    }
    default: { return 'cancel'
    }
  }
}

/**
 * Says how composing ended. It does not wait for the person to read it.
 * @param outcome - what `composeStack` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellComposeOutcome (outcome: ComposeOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'added': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(`MonoSmith: ${outcome.name} is added. ${outcome.committed ? 'Committed as one change: Undo reverts it.' : 'Not committed: review with git status.'}`)
      break
    }
    case 'recipe': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage('MonoSmith: the recipe is written. A recipe fixes the plugins, not the choices.')
      break
    }
    case 'copied': {
      void vscode.env.clipboard.writeText(outcome.command)
      void vscode.window.showInformationMessage('MonoSmith: the command is on the clipboard.')
      break
    }
    case 'refused': {
      void vscode.window.showErrorMessage(`MonoSmith: ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
