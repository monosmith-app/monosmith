import * as vscode from 'vscode'
import type { LevelDecider, LevelQuestion, RecipeBasics, RecipeChoice, RecipeOutcome, RecipeRequest } from './new-recipe.contract.ts'
import { checkChoiceAnswer, checkChoiceQuestion, checkRecipeName, checkRecipeVersion } from './recipe-levels.mapper.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.new-recipe.monosmith-new-recipe')
const UNSET = say('guide.new_recipe_levels.leave_unset_default')

/** The sentence that names what the recipe fixes. */
function fixedSentence (fixed: readonly string[]): string {
  return fixed.length > 0 ? say('vscode.new-recipe.the-recipe-fixes', { join: fixed.join('; ') }) : say('vscode.new-recipe.the-recipe-fixes-no-plugin')
}

/**
 * Asks the recipe's name, version and parent recipe.
 * @returns the answers, or undefined when the person pressed Escape at any question
 */
export async function askRecipeBasics (): Promise<RecipeBasics | undefined> {
  const name = await vscode.window.showInputBox({
    title:         say('vscode.new-recipe.what-is-the-recipe-called', { title: TITLE }),
    placeHolder:   'team',
    prompt:        say('vscode.new-recipe.lowercase-words-joined-by-dashes-the'),
    validateInput: checkRecipeName,
  })
  if (name === undefined) return undefined
  const version = await vscode.window.showInputBox({
    title:         `${TITLE}: version`,
    value:         '0.1.0',
    prompt:        say('vscode.new-recipe.major-minor-patch-enter-keeps-0'),
    validateInput: checkRecipeVersion,
  })
  if (version === undefined) return undefined
  const parent = await vscode.window.showInputBox({
    title:       say('vscode.new-recipe.recipe-it-extends', { title: TITLE }),
    placeHolder: say('vscode.new-recipe.base-recipe-toml-a-path-or'),
    prompt:      say('vscode.new-recipe.the-new-recipe-starts-from-that'),
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
  const fit = question.language === '' ? '' : say('vscode.new-recipe.that-fit', { language: question.language })
  const whenLeftOut = say('vscode.new-recipe.unset-the-recipe-it-extends-or')
  if (question.multi) {
    const chosen = await vscode.window.showQuickPick(plugins, {
      title:              say('vscode.new-recipe.which-plugins-does-the-recipe-fix', { title: TITLE, level: question.level }),
      placeHolder:        say('vscode.new-recipe.tick-the-plugins-tick-none-to', { fit, whenLeftOut }),
      canPickMany:        true,
      matchOnDescription: true,
    })

    return chosen?.map((item) => item.label)
  }
  const chosen = await vscode.window.showQuickPick([{ label: UNSET, description: say('vscode.new-recipe.no-plugin-is-fixed', { level: question.level }), unset: true }, ...plugins.map((item) => ({ ...item, unset: false }))], {
    title:              say('vscode.new-recipe.which-plugin-does-the-recipe-fix', { title: TITLE, level: question.level }),
    placeHolder:        say('vscode.new-recipe.pick-one', { fit, whenLeftOut }),
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
    { label: say('guide.new_recipe_chosen.fixed_by_the_recipe_default'), description: say('vscode.new-recipe.every-project-made-from-the-recipe', { level: question.level }), who: 'fixed' },
    { label: say('guide.new_recipe_chosen.chosen_by_the_people_who_use'), description: say('vscode.new-recipe.smith-asks-them-a-question-for', { level: question.level }), who: 'chosen' },
  ], { title: say('vscode.new-recipe.who-decides-the', { title: TITLE, level: question.level }), placeHolder: say('vscode.new-recipe.pick-who-decides') })

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
    title:         say('vscode.new-recipe.the-question-for-the', { title: TITLE, level: question.level }),
    value:         say('vscode.new-recipe.which-do-you-want', { level: question.level }),
    prompt:        say('vscode.new-recipe.the-people-who-make-a-project'),
    validateInput: checkChoiceQuestion,
  })
  if (text === undefined) return undefined
  if (question.multi) return await askGroupedAnswers(question, text.trim())
  const plugins = await vscode.window.showQuickPick(question.options.map((option) => ({ label: option.name, description: option.description })), {
    title:              say('vscode.new-recipe.which-plugins-can-they-end-up', { title: TITLE, level: question.level }),
    placeHolder:        say('vscode.new-recipe.tick-at-least-two-a-choice'),
    canPickMany:        true,
    matchOnDescription: true,
  })
  if (plugins === undefined) return undefined
  if (plugins.length < 2) {
    void vscode.window.showWarningMessage(say('vscode.new-recipe.monosmith-a-choice-needs-at-least'))

    return undefined
  }
  const answers: Array<{ answer: string, plugin: string }> = []
  for (const plugin of plugins) {
    const answer = await vscode.window.showInputBox({
      title:         say('vscode.new-recipe.the-answer-that-gives', { title: TITLE, label: plugin.label }),
      value:         plugin.label,
      prompt:        say('vscode.new-recipe.what-the-user-reads-for-this'),
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
      prompt:        say(enough ? 'vscode.new-recipe.name-another-answer-or-leave-it' : 'vscode.new-recipe.a-word-or-two-the-user'),
      validateInput: (value) => enough && value.trim() === '' ? undefined : checkChoiceAnswer(value, answers.map((a) => a.answer)),
    })
    if (answer === undefined) return undefined
    if (answer.trim() === '') return { question: text, answers }
    const plugins = await vscode.window.showQuickPick(question.options.map((option) => ({ label: option.name, description: option.description })), {
      title:              say('vscode.new-recipe.the-plugins-that-gives', { title: TITLE, trim: answer.trim() }),
      placeHolder:        say('vscode.new-recipe.tick-one-or-more-several-are'),
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
  const write = say('vscode.new-recipe.write-the-recipe')
  const fixed = [
    ...Object.entries(request.fixed).map(([level, plugins]) => `${level}: ${plugins.join(', ')}`),
    ...Object.entries(request.chosen ?? {}).map(([level, choice]) => say('vscode.new-recipe.chosen-by-the-users', { level, join: choice.answers.map((a) => a.answer).join(' or ') })),
  ]
  const choice = await vscode.window.showInformationMessage(
    say('vscode.new-recipe.monosmith-write', { file }),
    { modal: true, detail: say('vscode.new-recipe.nothing-was-written-yet-smith-checks', { command, part: fixedSentence(fixed) }) },
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
      const open = say('vscode.new-recipe.open-the-recipe')
      void vscode.window.showInformationMessage(say('vscode.new-recipe.monosmith-wrote-check-it-any-time', { file: outcome.file, file2: outcome.file }), open).then(async (choice) => {
        if (choice === open) await vscode.window.showTextDocument(vscode.Uri.joinPath(vscode.Uri.file(root), outcome.file))
      })
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.new-recipe.monosmith-no-recipe-was-written', { message: outcome.message }))
      break
    }
    case 'cancelled': {
      break
    }
  }
}
