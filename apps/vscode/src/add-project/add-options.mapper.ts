import type { ChoiceOption, KindOption, StackOption } from './add-options.contract.ts'

function records (json: string): Array<Record<string, unknown>> {
  const value: unknown = JSON.parse(json)
  if (!Array.isArray(value)) throw new TypeError('smith printed no list')

  return value.filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null)
}

function text (value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function words (value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : []
}

/**
 * Reads `smith kinds --json`.
 * @param json - what smith printed
 * @returns the kinds that have a name and a language
 * @throws SyntaxError when the text is not JSON, TypeError when it is not a list
 */
export function parseKinds (json: string): KindOption[] {
  return records(json)
    .map((item) => ({ kind: text(item.kind), language: text(item.language), label: text(item.label), description: text(item.description) }))
    .filter((item) => item.kind !== '' && item.language !== '')
}

/**
 * Reads `smith choices --json`.
 * @param json - what smith printed
 * @returns the choices that have an axis and a name
 * @throws SyntaxError when the text is not JSON, TypeError when it is not a list
 */
export function parseChoices (json: string): ChoiceOption[] {
  return records(json)
    .map((item) => ({
      axis:        text(item.axis),
      name:        text(item.name),
      description: text(item.description),
      ...(text(item.language) !== '' && { language: text(item.language) }),
      requires:    words(item.requires),
      isDefault:   item.default === true,
      many:        item.many === true,
      needs:       words(item.needs),
      suggested:   item.suggested === true,
      conflicts:   words(item.conflicts),
    }))
    .filter((item) => item.axis !== '' && item.name !== '')
}

/**
 * Reads `smith stacks --json`.
 * @param json - what smith printed
 * @returns the stacks that have a name
 * @throws SyntaxError when the text is not JSON, TypeError when it is not a list
 */
export function parseStacks (json: string): StackOption[] {
  return records(json)
    .map((item) => ({ name: text(item.name), description: text(item.description), language: text(item.language) }))
    .filter((item) => item.name !== '')
}

/**
 * What a combined kind adds to its language: `node-app` of `node` is `app`.
 * @param kind - the combined kind
 * @param language - its language
 */
export function roleOf (kind: string, language: string): string {
  return kind.startsWith(`${language}-`) ? kind.slice(language.length + 1) : kind
}

/**
 * The names of the choices a language offers (react, vue, fastapi), to show
 * under the language and to find it when the list is filtered by one of them.
 * @param choices - every choice smith lists
 * @param language - the language of the new project
 * @returns the names, without repeats, in the order smith lists them
 */
export function choiceNames (choices: readonly ChoiceOption[], language: string): string[] {
  return [...new Set(choicesByAxis(choices, language).flatMap((axis) => axis.choices.map((choice) => choice.name)))]
}

/**
 * The axes of a language that a choice needs and that have no default: the
 * trigger and the runtime of a function app. smith never picks them, so the
 * wizard asks each once a choice that needs it is picked (ADR 0072).
 * @param choices - every choice smith lists
 * @param language - the language of the new project
 * @returns the axis names, in the order smith lists them
 */
export function askedAxes (choices: readonly ChoiceOption[], language: string): string[] {
  const axes = choicesByAxis(choices, language)
  const defaulted = new Set(axes.filter((axis) => axis.choices.some((choice) => choice.isDefault)).map((axis) => axis.axis))
  const listed = new Set(axes.map((axis) => axis.axis))

  return [...new Set(axes.flatMap((axis) => axis.choices.flatMap((choice) => choice.needs)))].filter((name) => listed.has(name) && !defaulted.has(name))
}

/**
 * What a language's function hosts are called in words ("a serverless Azure
 * Functions app"), so that the language is found when the list is filtered by
 * "serverless", "google" or "cloud run". A host is a choice that needs an asked axis.
 * @param choices - every choice smith lists
 * @param language - the language of the new project
 * @returns the descriptions of the choices that need an asked axis
 */
export function hostTerms (choices: readonly ChoiceOption[], language: string): string[] {
  const asked = new Set(askedAxes(choices, language))

  return choicesByAxis(choices, language).flatMap((axis) => axis.choices.filter((choice) => choice.needs.some((need) => asked.has(need))).map((choice) => choice.description))
}

/**
 * The choices a project of a language can be built from, grouped by axis.
 * @param choices - every choice smith lists
 * @param language - the language of the new project
 * @returns the axes in the order smith lists them, each with its choices
 */
export function choicesByAxis (choices: readonly ChoiceOption[], language: string): Array<{ axis: string, many: boolean, choices: ChoiceOption[] }> {
  const axes = new Map<string, { axis: string, many: boolean, choices: ChoiceOption[] }>()
  for (const choice of choices) {
    if (choice.language !== undefined && choice.language !== language) continue
    const axis = axes.get(choice.axis) ?? { axis: choice.axis, many: false, choices: [] }
    axis.choices.push(choice)
    axis.many ||= choice.many
    axes.set(choice.axis, axis)
  }

  return Array.from(axes, ([, axis]) => axis)
}
