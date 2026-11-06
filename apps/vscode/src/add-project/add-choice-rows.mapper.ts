import type { AddOptions, AddRequest, ChoiceOption } from './add-options.contract.ts'
import { askedAxes, choicesByAxis } from './add-options.mapper.ts'
import { wordsOf, type OutcomeItem } from './add-outcomes.mapper.ts'
import { say } from '../words/index.ts'

/** One line of a single-choice list. An empty `name` is the line that leaves the axis out. */
export interface ChoiceRow {
  name:        string
  label:       string
  description: string
  detail:      string
}

/**
 * Whether a choice can be offered for what is being built. A cloud host fits
 * only a function app, and a desktop shell that sits among the frameworks
 * fits only a desktop app: a website is not an AWS Lambda.
 * @param choice - the choice
 * @param outcomeId - what is being built (`web`, `function`, `desktop`...); undefined on the long way, which offers all
 * @returns false when the choice conflicts with the outcome
 */
export function fitsOutcome (choice: ChoiceOption, outcomeId: string | undefined): boolean {
  if (choice.layer === 'host') return outcomeId === undefined || outcomeId === 'function'

  const wrapsAnApp = choice.layer === 'shell' && choice.axis !== 'shell'

  return !wrapsAnApp || outcomeId === undefined || outcomeId === 'desktop'
}

/**
 * The lines of a single-choice list, with the answer that Enter takes first:
 * the default of the axis when it has one, else "None" (an axis with no default
 * adds nothing when it is left out). A choice that does not go with what is
 * already picked, or with what is being built, is not listed.
 * @param axis - the axis and its choices
 * @param picked - the choices picked so far
 * @param outcomeId - what is being built; undefined on the long way
 * @param ofLayers - the title of each layer, to word "None"
 * @returns the lines, the first one being what Enter takes
 */
export function singleChoiceRows (axis: { axis: string, choices: readonly ChoiceOption[] }, picked: readonly ChoiceOption[], outcomeId: string | undefined, ofLayers: Readonly<Record<string, string>> = {}): ChoiceRow[] {
  const fits = axis.choices
    .filter((choice) => fitsOutcome(choice, outcomeId))
    .filter((choice) => picked.every((other) => !choice.conflicts.includes(other.name) && !other.conflicts.includes(choice.name)))
  const rows = fits.map((choice): ChoiceRow => ({
    name:        choice.name,
    label:       choice.title ?? choice.name,
    description: `${choice.isDefault ? say('vscode.add-project.default') : ''}${choice.requires.length > 0 ? say('vscode.add-project.needs', { join: choice.requires.join(', ') }) : ''}`.trim(),
    detail:      choice.description,
  }))
  const defaulted = rows.findIndex((_, index) => fits[index]?.isDefault)
  if (defaulted !== -1) return [...rows.slice(defaulted, defaulted + 1), ...rows.slice(0, defaulted), ...rows.slice(defaulted + 1)]
  const layer = (ofLayers[axis.axis] ?? axis.axis).toLowerCase()

  return [{ name: '', label: say('vscode.add-project.none'), description: say('vscode.add-project.default-nothing-extra'), detail: say('vscode.add-project.leave-out-the-you-can-add', { layer }) }, ...rows]
}

/**
 * The lines of the question about an axis that smith never picks (a trigger, a
 * runtime): the choices that go with what is picked, the suggested one first.
 * @param choices - the choices of the axis
 * @param picked - the choices picked so far
 * @returns the lines; the first is the suggested answer
 */
export function askedRows (choices: readonly ChoiceOption[], picked: readonly ChoiceOption[]): ChoiceRow[] {
  return choices
    .filter((choice) => picked.every((other) => !choice.conflicts.includes(other.name) && !other.conflicts.includes(choice.name)))
    .sort((a, b) => Number(b.suggested) - Number(a.suggested))
    .map((choice) => ({ name: choice.name, label: choice.name, description: choice.suggested ? '(suggested)' : '', detail: choice.description }))
}

/**
 * The item of the layer that what the person typed in the first question points
 * at (`react` points at React), so that the next list starts on it and not on
 * the first row.
 * @param items - the frameworks, shells or hosts of the outcome
 * @param typed - what was typed in the first question
 * @returns the name of the first item that every typed word starts a word of; undefined when nothing was typed or nothing matches
 */
export function typedItem (items: readonly OutcomeItem[], typed: string): string | undefined {
  const query = wordsOf(typed)
  if (query.length === 0) return undefined

  return items.find((item) => {
    const own = [item.name, item.title, ...item.aliases].flatMap((word) => wordsOf(word))

    return query.every((word) => own.some((candidate) => candidate.startsWith(word)))
  })?.name
}

/**
 * Says the choices of a request in words, and which of them the catalogue
 * marks as the default, so that a person who pressed Enter through the
 * questions can see what they got ("React, ESLint (default), Vitest (default);
 * no desktop shell, no style").
 * @param request - what the person asked for
 * @param options - what the questions offered
 * @returns one or two sentences; says so when nothing extra was chosen
 */
export function choicesInWords (request: AddRequest, options: AddOptions): string {
  if (request.stack !== undefined) return say('vscode.add-project.built-from-the-stack', { stack: request.stack })
  const axes = choicesByAxis(options.choices, request.language)
  const named = request.with.map((name) => axes.flatMap((axis) => axis.choices).find((choice) => choice.name === name)).filter((choice): choice is ChoiceOption => choice !== undefined)
  if (named.length === 0) return say('vscode.add-project.built-with-the-default-choices-of', { language: request.language })
  const asked = askedAxes(options.choices, request.language)
  const taken = new Set(named.map((choice) => choice.axis))
  const without = axes
    .filter((axis) => !axis.many && !asked.includes(axis.axis) && !taken.has(axis.axis) && axis.choices.every((choice) => !choice.isDefault))
    .map((axis) => `no ${(options.layers[axis.axis] ?? axis.axis).toLowerCase()}`)
  const chosen = named.map((choice) => `${choice.title ?? choice.name}${choice.isDefault ? say('vscode.add-project.default-2') : ''}`).join(', ')

  return say('vscode.add-project.your-choices', { chosen, part: without.length > 0 ? `; ${without.join(', ')}` : '' })
}
