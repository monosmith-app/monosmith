import type { KindOption } from './add-options.contract.ts'
import { roleOf } from './add-options.mapper.ts'
import { say } from '../words/index.ts'

/** What the words of a layer mean, the same sentences `smith compose` shows. */
const LAYER_ABOUT: Record<string, string> = {
  framework: say('vscode.add-project.how-the-app-or-library-is'),
  shell:     say('vscode.add-project.what-wraps-the-app-as-a'),
  host:      say('vscode.add-project.where-the-code-runs-in-the'),
  test:      say('vscode.add-project.the-tool-that-runs-the-tests'),
  lint:      say('vscode.add-project.the-tool-that-checks-and-formats'),
  style:     say('vscode.add-project.how-the-app-looks-a-css'),
}

/**
 * One line that says what a layer or an axis of choices means, to show where the
 * word first appears.
 * @param layer - framework, shell, host, test, lint or style
 * @returns "framework = how the app…", or '' for a word with no line
 */
export function whatIs (layer: string): string {
  const about = LAYER_ABOUT[layer]

  return about === undefined ? '' : `${layer} = ${about}.`
}

/** A framework, a desktop shell or a cloud host that names an outcome. */
export interface OutcomeItem {
  name:        string
  title:       string
  description: string
  language:    string
  /** `framework`, `shell` or `host`. */
  layer:       string
  aliases:     string[]
}

/** Something a person wants to build, in plain words: the first question of Add project. */
export interface OutcomeOption {
  /** `web`, `api`, `function`, `desktop`, `library` or `package`. */
  id:          string
  title:       string
  /** The kind smith adds: `app`, `lib` or `package`. */
  role:        string
  /** The languages that fit it, in the order smith lists them. One means no language is asked. */
  languages:   string[]
  /** What people usually pick for it, for the line of the list. */
  typical:     string
  /** One sentence: what the person gets. */
  consequence: string
  /** The words that find it besides its title. */
  words:       string[]
  /** The frameworks, shells or hosts of its layer, to pick one. */
  items:       OutcomeItem[]
}

/** A language smith can generate, with the words a person knows it by. */
export interface LanguageInfo {
  name:        string
  /** "C# / .NET". */
  title:       string
  description: string
}

type Json = Record<string, unknown>

function text (value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function words (value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string' && item !== '') : []
}

function records (value: unknown): Json[] {
  return Array.isArray(value) ? value.filter((item): item is Json => typeof item === 'object' && item !== null && !Array.isArray(item)) : []
}

/**
 * Reads what the first question of Add offers from `smith compose --json`: the
 * outcomes smith builds from its catalogue (the same list the terminal screen
 * shows) and the languages with their names.
 * @param json - what `smith compose --json` printed
 * @returns the outcomes, the languages and the title of each layer (`shell` is "Desktop shell"); empty when smith is too old to list them
 * @throws SyntaxError when the text is not JSON, TypeError when it is not an object
 */
export function parseOutcomes (json: string): { outcomes: OutcomeOption[], languages: LanguageInfo[], layers: Record<string, string> } {
  const value: unknown = JSON.parse(json)
  if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new TypeError(say('vscode.add-project.smith-printed-no-catalogue'))
  const root = value as Json
  const outcomes = records(root.outcomes).map((outcome): OutcomeOption => ({
    id:          text(outcome.id),
    title:       text(outcome.title),
    role:        text(outcome.kind),
    languages:   words(outcome.languages),
    typical:     words(outcome.typical).join(', '),
    consequence: text(outcome.consequence),
    words:       words(outcome.words),
    items:       records(outcome.items).map((item) => ({
      name:        text(item.name),
      title:       text(item.title) || text(item.name),
      description: text(item.description),
      language:    text(item.language),
      layer:       text(item.layer),
      aliases:     words(item.aliases),
    })),
  })).filter((outcome) => outcome.id !== '' && outcome.title !== '')
  const languages = records(root.items)
    .filter((item) => item.kind === 'language' && text(item.name) !== '')
    .map((item) => ({ name: text(item.name), title: text(item.title) || text(item.name), description: text(item.description) }))

  const layers = Object.fromEntries(records(root.layers).filter((layer) => text(layer.id) !== '' && text(layer.title) !== '').map((layer) => [text(layer.id), text(layer.title)]))

  return { outcomes, languages, layers }
}

/**
 * Keeps the languages of each outcome that have a kind of the outcome's role (a
 * library needs a lib kind), and leaves out an outcome with no language left.
 * @param outcomes - what smith lists
 * @param kinds - what `smith kinds` lists
 * @returns the outcomes smith can add here
 */
export function fitOutcomes (outcomes: readonly OutcomeOption[], kinds: readonly KindOption[]): OutcomeOption[] {
  const fitted: OutcomeOption[] = []
  for (const outcome of outcomes) {
    const languages = outcome.languages.filter((language) => kinds.some((kind) => kind.language === language && roleOf(kind.kind, kind.language) === outcome.role))
    if (languages.length > 0) fitted.push({ ...outcome, languages, items: outcome.items.filter((item) => languages.includes(item.language)) })
  }

  return fitted
}

/** One line of the first question. */
export type OutcomeRow =
  | { kind: 'outcome', outcome: OutcomeOption } |
  { kind: 'none', query: string, closest: string[] } |
  { kind: 'every' } |
  { kind: 'compose' }

/** Splits a text into lowercase words, the way the search of the first question does. */
export function wordsOf (value: string): string[] {
  return value.toLowerCase().split(/[^\p{L}\p{N}#+.]+/u).filter(Boolean)
}

/** The words that find an outcome, strongest first: its title, then its own words, then what it names. */
function haystack (outcome: OutcomeOption): Array<{ words: string[], weight: number }> {
  return [
    { words: wordsOf(outcome.title), weight: 3 },
    { words: outcome.words.flatMap((word) => wordsOf(word)), weight: 2 },
    { words: [...outcome.items.flatMap((item) => [item.name, item.title, ...item.aliases]), outcome.typical].flatMap((word) => wordsOf(word)), weight: 2 },
    { words: wordsOf(outcome.consequence), weight: 1 },
  ]
}

function score (outcome: OutcomeOption, query: readonly string[]): number {
  let total = 0
  for (const word of query) {
    const found = haystack(outcome).filter((part) => part.words.some((candidate) => candidate.startsWith(word))).map((part) => part.weight)
    if (found.length === 0) return 0
    total += Math.max(...found)
  }

  return total
}

function distance (a: string, b: string): number {
  let previous = Array.from({ length: b.length + 1 }, (_, index) => index)
  for (let i = 1; i <= a.length; i++) {
    const row = [i]
    for (let j = 1; j <= b.length; j++) row[j] = Math.min((previous[j] ?? 0) + 1, (row[j - 1] ?? 0) + 1, (previous[j - 1] ?? 0) + (a[i - 1] === b[j - 1] ? 0 : 1))
    previous = row
  }

  return previous[b.length] ?? 0
}

/**
 * The outcomes that a search finds, the same way the terminal screen searches: every
 * word of the search must start a word of the title, of the outcome's own words,
 * or of the name, title or alias of a framework, shell or host it offers.
 * When nothing matches, one row says so and names the closest outcomes. The
 * rows that lead elsewhere (every language, compose) always stay.
 * @param outcomes - what smith can add here
 * @param query - what the person typed
 * @returns the rows of the list, best match first
 */
export function outcomeRows (outcomes: readonly OutcomeOption[], query: string): OutcomeRow[] {
  const typed = wordsOf(query)
  const tail: OutcomeRow[] = [{ kind: 'every' }, { kind: 'compose' }]
  if (typed.length === 0) return [...outcomes.map((outcome): OutcomeRow => ({ kind: 'outcome', outcome })), ...tail]

  const found = outcomes
    .map((outcome, index) => ({ outcome, index, points: score(outcome, typed) }))
    .filter((entry) => entry.points > 0)
    .sort((a, b) => b.points - a.points || a.index - b.index)
  if (found.length > 0) return [...found.map((entry): OutcomeRow => ({ kind: 'outcome', outcome: entry.outcome })), ...tail]

  const closest = outcomes
    .map((outcome) => ({ outcome, far: Math.min(...typed.flatMap((word) => haystack(outcome).flatMap((part) => part.words.map((candidate) => distance(word, candidate))))) }))
    .sort((a, b) => a.far - b.far)
    .slice(0, 3)
    .map((entry) => entry.outcome.title)

  return [{ kind: 'none', query: query.trim(), closest }, ...tail]
}

/** One row of the language question. */
export interface LanguageRow {
  /** The name `--lang` takes. */
  language:    string
  /** "C# / .NET". */
  label:       string
  /** What the language is for here: the clouds of a function app, the shells of a desktop app. */
  description: string
  detail:      string
}

/**
 * The rows of the language question in words a person knows: the language by
 * its name, and next to it what it offers for the thing being built (the clouds
 * of a function app, the shells of a desktop app), so no list of ids shows.
 * @param languages - the languages smith lists, with their names
 * @param kinds - what `smith kinds` lists
 * @param role - `app`, `lib` or `package`
 * @param fitting - the languages that fit what is being built
 * @param outcome - what is being built, if the person chose one
 * @returns one row per language, in the order of `fitting`
 */
export function languageRows (languages: readonly LanguageInfo[], kinds: readonly KindOption[], role: string, fitting: readonly string[], outcome?: OutcomeOption): LanguageRow[] {
  const rows: LanguageRow[] = []
  for (const name of fitting) {
    const kind = kinds.find((candidate) => candidate.language === name && roleOf(candidate.kind, candidate.language) === role)
    if (kind === undefined) continue
    const info = languages.find((language) => language.name === name)
    const offers = [...new Set((outcome?.items ?? []).filter((item) => item.language === name).map((item) => item.title))]
    rows.push({ language: name, label: info?.title ?? name, description: offers.join(', '), detail: info?.description || kind.description || kind.label })
  }

  return rows
}
