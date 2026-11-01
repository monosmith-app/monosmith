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
