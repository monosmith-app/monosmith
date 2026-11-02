import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import type { LevelDecider, LevelQuestion, RecipeBasics, RecipeChoice, RecipeOutcome, RecipeRequest } from './new-recipe.contract.ts'
import { parseRecipeLevels, questionFor, recipeArguments, recipeFile } from './recipe-levels.mapper.ts'

/** What writing a recipe needs from the editor. */
export interface NewRecipeDeps {
  /** Runs smith in the workspace folder for a short answer (the list of levels). */
  run:       (args: string[]) => Promise<SmithRun>
  /** Runs smith in the workspace folder to the end (it writes the file). */
  apply:     (args: string[]) => Promise<SmithRun>
  /** Asks the name, the version and the parent recipe. Resolves undefined when the person cancels. */
  askBasics: () => Promise<RecipeBasics | undefined>
  /** Asks one level: one plugin, or any number. Resolves the names picked (empty leaves the level unset), or undefined when the person cancels. */
  pick:      (question: LevelQuestion) => Promise<string[] | undefined>
  /** Asks, for a level that takes one plugin, who decides it: the recipe (default) or the people who use it. Resolves undefined when the person cancels. */
  decide:    (question: LevelQuestion) => Promise<LevelDecider | undefined>
  /** Asks the question the users read and the plugin each answer gives. Resolves undefined when the person cancels. */
  askChoice: (question: LevelQuestion) => Promise<RecipeChoice | undefined>
  /** Shows the command and the file it writes, and asks. */
  confirm:   (request: RecipeRequest, command: string, file: string) => Promise<boolean>
}

/**
 * Writes a recipe. It reads the levels and plugins from smith (no name is
 * held here), asks the basics and, for each level, who decides it: the recipe
 * fixes a plugin (the default), or the people who use the recipe choose by
 * answering a question that each maps to a plugin. It shows the exact
 * `smith recipe new` command and the file it writes, and only on a yes runs it.
 * A level that takes any number of plugins is fixed only.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function newRecipe (deps: NewRecipeDeps): Promise<RecipeOutcome> {
  const listed = await deps.run(['recipe', 'new', '--list-levels', '--json'])
  if (listed.code !== 0) return { kind: 'refused', message: describeFailure(listed) }
  let levels
  try {
    levels = parseRecipeLevels(listed.stdout)
  } catch {
    return { kind: 'refused', message: 'smith printed its levels in a form this editor does not read. Update the extension, or run smith recipe new in a terminal.' }
  }
  if (levels.levels.length === 0) return { kind: 'refused', message: 'smith knows no plugin a recipe can fix. Update smith, or run smith recipe new in a terminal.' }

  const basics = await deps.askBasics()
  if (basics === undefined) return { kind: 'cancelled' }
  const fixed: Record<string, string[]> = {}
  const chosen: Record<string, RecipeChoice> = {}
  let language = ''
  for (const level of levels.levels) {
    const question = questionFor(levels, level, language)
    if (question === undefined) continue
    if (question.options.length >= 2) {
      const who = await deps.decide(question)
      if (who === undefined) return { kind: 'cancelled' }
      if (who === 'chosen') {
        const choice = await deps.askChoice(question)
        if (choice === undefined) return { kind: 'cancelled' }
        chosen[level.level] = choice
        if (level.level === 'language') language = ''
        continue
      }
    }
    const picked = await deps.pick(question)
    if (picked === undefined) return { kind: 'cancelled' }
    if (picked.length > 0) fixed[level.level] = picked
    if (level.level === 'language') language = picked[0] ?? ''
  }

  const request: RecipeRequest = { ...basics, fixed, chosen }
  const args = recipeArguments(request)
  const command = ['smith', ...args.map((arg) => /\s/.test(arg) ? `"${arg}"` : arg)].join(' ')
  if (!await deps.confirm(request, command, recipeFile(request))) return { kind: 'cancelled' }
  const done = await deps.apply(args)

  return done.code === 0
    ? { kind: 'written', request, file: recipeFile(request), lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) }
    : { kind: 'refused', message: describeFailure(done) }
}
