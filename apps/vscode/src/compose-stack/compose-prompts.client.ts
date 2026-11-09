import * as vscode from 'vscode'
import { checkProjectName } from '../add-project/index.ts'
import { tellDoctorFailure, type DoctorFailureFollowUp } from '../doctor-failure/index.ts'
import { pickLabels, settleRivals, tickedTitles } from './compose-report.mapper.ts'
import type { ComposeAnswer, ComposeItem, ComposeOption, ComposeOutcome, ComposeQuestion } from './compose-stack.contract.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.compose-stack.monosmith-compose-a-stack')

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
  type Entry = vscode.QuickPickItem & { id?: string }
  const pick = vscode.window.createQuickPick<Entry>()
  pick.canSelectMany = true
  pick.matchOnDescription = true
  pick.matchOnDetail = true
  pick.ignoreFocusOut = true
  pick.placeholder = say('vscode.compose-stack.search-then-tick-what-you-want')
  let current = [...ticked]
  let updating = false

  // A ticked row is always shown, whatever the search says, so what is chosen never disappears.
  const render = (): void => {
    const entries: Entry[] = []
    let layer = ''
    for (const item of items) {
      if (item.layer !== layer) {
        layer = item.layer
        entries.push({ label: layer, kind: vscode.QuickPickItemKind.Separator })
      }
      entries.push({ ...pickLabels(item), id: item.id, alwaysShow: current.includes(item.id) })
    }
    // Rebuilding the list moves the cursor to the first row: put it back, so the arrow keys carry on from where they were.
    const active = new Set(pick.activeItems.map((entry) => entry.id))
    updating = true
    pick.items = entries
    pick.selectedItems = entries.filter((entry) => entry.id !== undefined && current.includes(entry.id))
    pick.activeItems = entries.filter((entry) => entry.id !== undefined && active.has(entry.id))
    updating = false
    setTitle()
  }
  const setTitle = (): void => {
    const chosen = tickedTitles(items, current)
    pick.title = `${TITLE} (1/3): ${message === '' ? say('vscode.compose-stack.what-do-you-want') : message}${chosen === '' ? '' : say('vscode.compose-stack.selected', { chosen })}`
  }
  render()

  return await new Promise((resolve) => {
    let done = false
    pick.onDidChangeSelection((selection) => {
      if (updating) return
      const wanted = selection.map((entry) => entry.id).filter((id): id is string => id !== undefined)
      const next = settleRivals(items, current, wanted).ids
      const changed = next.length !== wanted.length
      current = next
      if (changed) {
        render()
      } else {
        setTitle()
      }
    })
    pick.onDidAccept(() => {
      done = true
      resolve(current)
      pick.hide()
    })
    pick.onDidHide(() => {
      if (!done) resolve(undefined)
      pick.dispose()
    })
    pick.show()
  })
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
    options.map((option) => ({ label: `${option.title}${option.suggested ? say('vscode.compose-stack.suggested') : ''}`, detail: option.description, option })),
    { title: `${TITLE}: which ${question.axis} for ${question.choice}?`, placeHolder: say('vscode.compose-stack.smith-does-not-choose-this-for'), matchOnDetail: true },
  )

  return picked?.option
}

/**
 * Asks the project name. The name comes before the plan because the files
 * of the plan are named after it (apps/<name>/); the prompt says so.
 * @param existing - names the workspace has already
 * @param stack - the stack in words ("React + Electron (node)")
 * @returns the name, or undefined when the person pressed Escape
 */
export async function askProjectName (existing: readonly string[], stack = ''): Promise<string | undefined> {
  return await vscode.window.showInputBox({
    title:         say('vscode.compose-stack.2-3-name-your', { title: TITLE, part: stack === '' ? say('vscode.compose-stack.app') : stack }),
    prompt:        say('vscode.compose-stack.smith-shows-the-whole-plan-next'),
    placeHolder:   say('vscode.compose-stack.for-example-web'),
    validateInput: (value) => checkProjectName(value) ?? (existing.includes(value) ? say('vscode.add-project.exists-already-pick-another-name', { value }) : undefined),
  })
}

/**
 * Shows the plan and asks what to do with it. Nothing was changed yet.
 * @param plan - the words of the plan
 * @param hasRecipe - whether the plugins can be saved as a recipe
 * @param toolsMissing - whether a tool of the plan is missing, which adds the button that installs it
 * @returns the answer
 */
export async function confirmComposePlan (plan: string, hasRecipe: boolean, toolsMissing = false): Promise<ComposeAnswer> {
  const run = say('vscode.compose-stack.run-it')
  const commit = say('vscode.compose-stack.run-it-and-commit')
  const copy = say('vscode.compose-stack.copy-the-command')
  const recipe = say('vscode.compose-stack.save-the-plugins-as-a-recipe')
  const tools = say('vscode.compose-stack.install-the-tools-first')
  const missing = toolsMissing ? say('vscode.compose-stack.a-tool-above-is-missing-you') : ''
  const choice = await vscode.window.showInformationMessage(
    say('vscode.compose-stack.monosmith-3-3-this-is-the'),
    { modal: true, detail: say('vscode.compose-stack.nothing-was-changed-yet-by-default', { plan, missing }) },
    run,
    commit,
    ...(toolsMissing ? [tools] : []),
    copy,
    ...(hasRecipe ? [recipe] : []),
  )
  switch (choice) {
    case tools: { return 'tools'
    }
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
 * @param next - carries out the buttons that need the editor (the commit after the fact)
 */
export function tellComposeOutcome (outcome: ComposeOutcome, onOutput: (line: string) => void, next?: DoctorFailureFollowUp): void {
  switch (outcome.kind) {
    case 'added': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(say('vscode.compose-stack.monosmith-is-added', { name: outcome.name, say: say(outcome.committed ? 'vscode.add-project.committed-as-one-change-undo-reverts' : 'vscode.compose-stack.not-committed-review-with-git-status') }))
      break
    }
    case 'written': {
      tellDoctorFailure(outcome.failure, outcome.operation, outcome.lines, onOutput, next)
      break
    }
    case 'recipe': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(say('vscode.compose-stack.monosmith-the-recipe-is-written-a'))
      break
    }
    case 'copied': {
      void vscode.env.clipboard.writeText(outcome.command)
      void vscode.window.showInformationMessage(say('vscode.compose-stack.monosmith-the-command-is-on-the'))
      break
    }
    case 'refused': {
      void vscode.window.showErrorMessage(say('vscode.compose-stack.monosmith', { message: outcome.message }))
      break
    }
    case 'cancelled': {
      break
    }
  }
}
