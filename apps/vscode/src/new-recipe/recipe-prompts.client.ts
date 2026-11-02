import * as vscode from 'vscode'
import type { LevelDecider, LevelQuestion, RecipeBasics, RecipeChoice, RecipeOutcome, RecipeRequest } from './new-recipe.contract.ts'
import { checkChoiceAnswer, checkChoiceQuestion, checkRecipeName, checkRecipeVersion } from './recipe-levels.mapper.ts'

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
  const whenLeftOut = 'Unset: the recipe it extends, or the person who uses the recipe, decides.'
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
 * Asks who decides a level that takes one plugin. "Fixed by the recipe" is the
 * default and comes first; each answer says what it means for the projects.
 * @param question - the level and the plugins that can fill it
 * @returns who decides, or undefined when the person pressed Escape
 */
export async function decideLevel (question: LevelQuestion): Promise<LevelDecider | undefined> {
  const picked = await vscode.window.showQuickPick<{ label: string, description: string, who: LevelDecider }>([
    { label: 'Fixed by the recipe (default)', description: `every project made from the recipe gets the ${question.level} plugin you pick next, or none`, who: 'fixed' },
    { label: 'Chosen by the people who use it', description: `smith asks them a question for the ${question.level} when they make a project; each answer gives a plugin`, who: 'chosen' },
  ], { title: `${TITLE}: who decides the ${question.level}?`, placeHolder: 'Pick who decides' })

  return picked?.who
}

/**
 * Asks the question the people who use the recipe read, the plugins they can
 * end up with, and the name of each answer (it starts on the plugin's name).
 * @param question - the level and the plugins that can fill it
 * @returns the choice, or undefined when the person pressed Escape at any step
 */
export async function askChoice (question: LevelQuestion): Promise<RecipeChoice | undefined> {
  const text = await vscode.window.showInputBox({
    title:         `${TITLE}: the question for the ${question.level}`,
    value:         `Which ${question.level} do you want?`,
    prompt:        'The people who make a project from the recipe read it. Do not use a colon.',
    validateInput: checkChoiceQuestion,
  })
  if (text === undefined) return undefined
  if (question.multi) return await askGroupedAnswers(question, text.trim())
  const plugins = await vscode.window.showQuickPick(question.options.map((option) => ({ label: option.name, description: option.description })), {
    title:              `${TITLE}: which ${question.level} plugins can they end up with?`,
    placeHolder:        'Tick at least two: a choice with one answer is a fixed plugin',
    canPickMany:        true,
    matchOnDescription: true,
  })
  if (plugins === undefined) return undefined
  if (plugins.length < 2) {
    void vscode.window.showWarningMessage('MonoSmith: a choice needs at least two plugins. Tick two or more, or fix the plugin instead.')

    return undefined
  }
  const answers: Array<{ answer: string, plugin: string }> = []
  for (const plugin of plugins) {
    const answer = await vscode.window.showInputBox({
      title:         `${TITLE}: the answer that gives ${plugin.label}`,
      value:         plugin.label,
      prompt:        'What the user reads for this plugin. Enter keeps the plugin name.',
      validateInput: (value) => checkChoiceAnswer(value, answers.map((a) => a.answer)),
    })
    if (answer === undefined) return undefined
    answers.push({ answer: answer.trim(), plugin: plugin.label })
  }

  return { question: text.trim(), answers }
}

/**
 * Asks the answers of a level that takes any number of plugins: each answer has a name and the
 * plugins it gives, ticked from the list (several plugins of one answer are joined by "+").
 * The person ends the list with an empty name, once at least two answers exist.
 * @param question - the level and the plugins that can fill it
 * @param text - the question the users read
 * @returns the choice, or undefined when the person pressed Escape
 */
async function askGroupedAnswers (question: LevelQuestion, text: string): Promise<RecipeChoice | undefined> {
  const answers: Array<{ answer: string, plugin: string }> = []
  for (;;) {
    const enough = answers.length >= 2
    const answer = await vscode.window.showInputBox({
      title:         `${TITLE}: answer ${answers.length + 1} of the ${question.level}`,
      prompt:        enough ? 'Name another answer, or leave it empty to finish.' : 'A word or two the user reads. A choice needs at least two answers.',
      validateInput: (value) => enough && value.trim() === '' ? undefined : checkChoiceAnswer(value, answers.map((a) => a.answer)),
    })
    if (answer === undefined) return undefined
    if (answer.trim() === '') return { question: text, answers }
    const plugins = await vscode.window.showQuickPick(question.options.map((option) => ({ label: option.name, description: option.description })), {
      title:              `${TITLE}: the plugins that "${answer.trim()}" gives`,
      placeHolder:        'Tick one or more; several are installed together',
      canPickMany:        true,
      matchOnDescription: true,
    })
    if (plugins === undefined) return undefined
    if (plugins.length === 0) continue
    answers.push({ answer: answer.trim(), plugin: plugins.map((item) => item.label).join('+') })
  }
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
  const fixed = [
    ...Object.entries(request.fixed).map(([level, plugins]) => `${level}: ${plugins.join(', ')}`),
    ...Object.entries(request.chosen ?? {}).map(([level, choice]) => `${level}: chosen by the users (${choice.answers.map((a) => a.answer).join(' or ')})`),
  ]
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
