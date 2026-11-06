import type { LevelQuestion, RecipeLevel, RecipeLevels, RecipePlugin, RecipeRequest } from './new-recipe.contract.ts'
import { say } from '../words/index.ts'

type Json = Record<string, unknown>

const NAME = /^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$/
const VERSION = /^\d+\.\d+\.\d+$/

function records (value: unknown): Json[] {
  return Array.isArray(value) ? value.filter((item): item is Json => typeof item === 'object' && item !== null && !Array.isArray(item)) : []
}

function text (value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function object (value: unknown): Json {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as Json : {}
}

/**
 * Reads `smith recipe new --list-levels --json`.
 * @param json - what smith printed
 * @returns the levels that have a name, and the plugins that fit each language
 * @throws SyntaxError when the text is not JSON, TypeError when it is not an object
 */
export function parseRecipeLevels (json: string): RecipeLevels {
  const value: unknown = JSON.parse(json)
  if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new TypeError(say('vscode.new-recipe.smith-printed-no-levels'))
  const root = value as Json
  const levels: RecipeLevel[] = records(root.levels)
    .map((l) => ({
      level:    text(l.level),
      additive: l.additive === true,
      plugins:  records(l.plugins).map((p) => ({ name: text(p.name), description: text(p.description) })).filter((p) => p.name !== ''),
    }))
    .filter((l) => l.level !== '')
  const fits: Record<string, Record<string, string[]>> = {}
  const byLanguage = Object.entries(object(root.fits))
  for (const [language, byLevel] of byLanguage) {
    const known: Record<string, string[]> = {}
    const perLevel = Object.entries(object(byLevel))
    for (const [level, names] of perLevel) {
      known[level] = Array.isArray(names) ? names.filter((n): n is string => typeof n === 'string') : []
    }
    fits[language] = known
  }

  return { levels, fits }
}

/**
 * The question to ask about a level, or undefined when no plugin can fill it.
 * After a language is fixed only the plugins that fit it are offered.
 * @param levels - what smith offers
 * @param level - the level to ask about
 * @param language - the language plugin fixed so far, or empty
 */
export function questionFor (levels: RecipeLevels, level: RecipeLevel, language: string): LevelQuestion | undefined {
  let options: RecipePlugin[] = level.plugins
  if (language !== '' && level.level !== 'language') {
    const names = levels.fits[language]?.[level.level] ?? []
    options = level.plugins.filter((p) => names.includes(p.name))
  }

  return options.length === 0 ? undefined : { level: level.level, multi: level.additive, language, options }
}

/**
 * The arguments of `smith recipe new` for a request: the name, the version and
 * the parent when given, and one `--fixed level=plugin` for each level (plugins
 * of a level that takes several are joined by commas).
 * @param request - what the person chose
 */
export function recipeArguments (request: RecipeRequest): string[] {
  const args = ['recipe', 'new', '--name', request.name]
  if (request.version !== '') args.push('--recipe-version', request.version)
  if (request.extends !== '') args.push('--extends', request.extends)
  for (const [level, plugins] of Object.entries(request.fixed)) {
    if (plugins.length > 0) args.push('--fixed', `${level}=${plugins.join(',')}`)
  }
  const chosen = Object.entries(request.chosen ?? {})
  for (const [level, choice] of chosen) {
    args.push('--chosen', `${level}=${choice.question}:${choice.answers.map((a) => `${a.answer}=${a.plugin}`).join(',')}`)
  }

  return args
}

/** Says what is wrong with the question the users read, or undefined when it is fine. */
export function checkChoiceQuestion (question: string): string | undefined {
  if (question.trim() === '') return say('vscode.new-recipe.write-the-question-the-people-who')

  return /[:\n]/.test(question) ? say('vscode.new-recipe.write-the-question-without-a-colon') : undefined
}

/**
 * Says what is wrong with the name of an answer, or undefined when it is fine.
 * @param answer - what the person typed
 * @param taken - the answers already named for this choice
 */
export function checkChoiceAnswer (answer: string, taken: string[]): string | undefined {
  if (answer.trim() === '' || /[:=+,]/.test(answer)) return say('vscode.new-recipe.an-answer-is-a-word-or')

  return taken.includes(answer.trim()) ? say('vscode.new-recipe.two-answers-are-named-each-needs', { trim: answer.trim() }) : undefined
}

/**
 * The file `smith recipe new` writes: it is named after the recipe, in the workspace folder.
 * @param request - what the person chose
 */
export function recipeFile (request: RecipeRequest): string {
  return `${request.name}.recipe.toml`
}

/** Says what is wrong with a recipe name, or undefined when it is fine. */
export function checkRecipeName (name: string): string | undefined {
  return NAME.test(name) ? undefined : say('vscode.new-recipe.use-lowercase-words-joined-by-dashes')
}

/** Says what is wrong with a version, or undefined when it is fine. Empty takes 0.1.0. */
export function checkRecipeVersion (version: string): string | undefined {
  return version === '' || VERSION.test(version) ? undefined : say('vscode.new-recipe.use-major-minor-patch-for-example')
}
