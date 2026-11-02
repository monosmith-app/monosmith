import * as vscode from 'vscode'
import type { LevelQuestion, RecipeBasics, RecipeOutcome, RecipeRequest } from './new-recipe.contract.ts'
import { checkRecipeName, checkRecipeVersion } from './recipe-levels.mapper.ts'

const TITLE = 'MonoSmith: New recipe'
const UNSET = 'Leave unset (default)'

/**
 * Asks the recipe's name, version and parent recipe.
 * @returns the answers, or undefined when the person pressed Escape at any question
 */
export async function askRecipeBasics (): Promise<RecipeBasics | undefined> {
  const name = await vscode.window.showInputBox({
    title:         `${TITLE}: what is the recipe called?`,
    placeHolder:   'team',
    prompt:        'Lowercase words joined by dashes. The file is <name>.recipe.toml.',
    validateInput: checkRecipeName,
  })
  if (name === undefined) return undefined
  const version = await vscode.window.showInputBox({
    title:         `${TITLE}: version`,
    value:         '0.1.0',
    prompt:        'major.minor.patch. Enter keeps 0.1.0.',
    validateInput: checkRecipeVersion,
  })
  if (version === undefined) return undefined
  const parent = await vscode.window.showInputBox({
    title:       `${TITLE}: recipe it extends`,
    placeHolder: 'base.recipe.toml (a path or an https URL; enter for none)',
    prompt:      'The new recipe starts from that one and changes what it names.',
  })

  return parent === undefined ? undefined : { name, version, extends: parent.trim() }
}

/**
 * Asks one level. A level that takes one plugin is a pick list that starts on
 * "Leave unset"; a level that takes any number is a list of tick boxes (nothing
 * ticked leaves it unset). Each plugin shows what it is.
 * @param question - the level and the plugins that can fill it
 * @returns the plugins picked (empty to leave the level unset), or undefined when the person pressed Escape
 */
export async function pickLevel (question: LevelQuestion): Promise<string[] | undefined> {
  const plugins = question.options.map((option) => ({ label: option.name, description: option.description }))
  const fit = question.language === '' ? '' : ` that fit ${question.language}`
  const whenLeftOut = 'Unset: the recipe it extends, or the person who uses the recipe, decides. To let users choose between plugins, run smith recipe new in a terminal.'
  if (question.multi) {
    const chosen = await vscode.window.showQuickPick(plugins, {
      title:              `${TITLE}: which ${question.level} plugins does the recipe fix?`,
      placeHolder:        `Tick the plugins${fit}; tick none to leave it unset. ${whenLeftOut}`,
      canPickMany:        true,
      matchOnDescription: true,
    })

    return chosen?.map((item) => item.label)
  }
  const chosen = await vscode.window.showQuickPick([{ label: UNSET, description: `no ${question.level} plugin is fixed`, unset: true }, ...plugins.map((item) => ({ ...item, unset: false }))], {
    title:              `${TITLE}: which ${question.level} plugin does the recipe fix?`,
    placeHolder:        `Pick one${fit}. ${whenLeftOut}`,
    matchOnDescription: true,
  })

  if (chosen === undefined) return undefined

  return chosen.unset ? [] : [chosen.label]
}

/**
 * Shows the exact command and the file it writes, and asks. Nothing is written before a yes.
 * @param request - what the person chose
 * @param command - the `smith recipe new` line that runs
 * @param file - the recipe file, in the workspace folder
 * @returns true to write the recipe
 */
export async function confirmRecipe (request: RecipeRequest, command: string, file: string): Promise<boolean> {
  const write = 'Write the recipe'
  const fixed = Object.entries(request.fixed).map(([level, plugins]) => `${level}: ${plugins.join(', ')}`)
  const choice = await vscode.window.showInformationMessage(
    `MonoSmith: write ${file}?`,
    { modal: true, detail: `${command}\n\n${fixed.length > 0 ? `The recipe fixes ${fixed.join('; ')}.` : 'The recipe fixes no plugin.'}\n\nNothing was written yet. smith checks that the recipe resolves before it writes the file, and does not replace a file that exists. The file is not committed.` },
    write,
  )

  return choice === write
}

/**
 * Says how writing the recipe ended, and offers to open the file. It does not wait for the person to read it.
 * @param outcome - what `newRecipe` returned
 * @param root - the workspace folder the file was written in
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellRecipeOutcome (outcome: RecipeOutcome, root: string, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'written': {
      for (const line of outcome.lines) onOutput(line)
      const open = 'Open the recipe'
      void vscode.window.showInformationMessage(`MonoSmith: wrote ${outcome.file}. Check it any time with smith recipe check ${outcome.file} --all.`, open).then(async (choice) => {
        if (choice === open) await vscode.window.showTextDocument(vscode.Uri.joinPath(vscode.Uri.file(root), outcome.file))
      })
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: no recipe was written. ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
