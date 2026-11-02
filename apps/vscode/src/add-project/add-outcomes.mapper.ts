import type { AddOptions, ChoiceOption } from './add-options.contract.ts'
import { roleOf } from './add-options.mapper.ts'

/** What the words of a layer mean, the same sentences `smith compose` shows. */
const LAYER_ABOUT: Record<string, string> = {
  framework: 'how the app or library is written: a web framework, or a whole starter',
  shell:     'what wraps the app as a program you install on a computer',
  host:      'where the code runs in the cloud, with the trigger and the runtime it asks for',
  test:      'the tool that runs the tests',
  lint:      'the tool that checks and formats the code',
  style:     'how the app looks: a CSS library',
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
  items:       ChoiceOption[]
}

interface Spec { id: string, title: string, role: string, layer?: string, host?: boolean, consequence: string, words: string[], typical?: string }

const BACKEND_LANGUAGES = ['node', 'python', 'go', 'dotnet', 'jvm', 'rust', 'php', 'ruby']

// The words of the product. What an outcome holds (languages, frameworks, hosts) comes from smith's choices, as in `smith compose`.
const SPECS: Spec[] = [
  { id: 'web', title: 'A website or web app', role: 'app', layer: 'framework', consequence: 'People open it in a browser. smith adds a dev server, tests and a build.', words: ['website', 'web app', 'web', 'frontend', 'page', 'browser', 'site'] },
  { id: 'api', title: 'An API or backend service', role: 'app', consequence: 'A program that answers requests on a server. You deploy it; it has no screen.', words: ['api', 'backend', 'server', 'service', 'rest', 'microservice', 'http'] },
  { id: 'function', title: 'A function app (Azure, AWS, Google)', role: 'app', layer: 'host', host: true, consequence: 'Small code that the cloud runs when something happens, such as a request or a timer.', words: ['function', 'functions', 'serverless', 'cloud', 'lambda', 'faas', 'trigger'] },
  { id: 'desktop', title: 'A desktop app (Tauri, Electron)', role: 'app', layer: 'shell', consequence: 'A program people install on a computer. smith adds the tasks that package it.', words: ['desktop', 'windows', 'mac', 'linux', 'native', 'installer'] },
  { id: 'library', title: 'A library for my other projects', role: 'lib', consequence: 'Shared code that the projects of this workspace import. It stays here and is not published.', words: ['library', 'lib', 'shared', 'internal', 'common', 'core', 'reuse'] },
  { id: 'package', title: 'A package I publish', role: 'package', typical: 'npm, PyPI, NuGet, crates.io', consequence: 'Code that other people install from a registry. smith adds the publish steps.', words: ['package', 'publish', 'npm', 'pypi', 'nuget', 'crates', 'registry', 'sdk'] },
]

/**
 * What a person can build, from what smith lists: the languages of a website, a
 * function app or a desktop app are those with a choice on its layer, so a new
 * plugin changes the list with no change here (the same rule as `smith compose`).
 * @param options - what `smith kinds` and `smith choices` list
 * @returns the outcomes in the order of the list; one with no language or no item on its layer is left out
 */
export function outcomesOf (options: AddOptions): OutcomeOption[] {
  const known = [...new Set(options.kinds.map((kind) => kind.language))]
  const result: OutcomeOption[] = []
  for (const spec of SPECS) {
    const items = spec.layer === undefined ? [] : options.choices.filter((choice) => (choice.layer ?? choice.axis) === spec.layer && (spec.host !== true || choice.needs.length > 0) && choice.language !== undefined && known.includes(choice.language))
    if (spec.layer !== undefined && items.length === 0) continue

    const languages = spec.layer === undefined
      ? (spec.id === 'api' ? BACKEND_LANGUAGES.filter((language) => known.includes(language)) : known)
      : [...new Set(items.map((item) => item.language).filter((language) => language !== undefined))]
    const role = spec.role
    const fitting = languages.filter((language) => options.kinds.some((kind) => kind.language === language && roleOf(kind.kind, kind.language) === role))
    if (fitting.length === 0) continue

    const typical = spec.typical ?? (spec.layer === undefined ? fitting.join(', ') : [...new Set(items.map((item) => item.title ?? item.name))].join(', '))
    result.push({
      id:          spec.id,
      title:       spec.title,
      role,
      languages:   fitting,
      typical,
      consequence: spec.consequence,
      items,
      words:       [...spec.words, ...items.flatMap((item) => [item.name, item.title ?? '', ...(item.aliases ?? [])]).filter((word) => word !== '')],
    })
  }

  return result
}
