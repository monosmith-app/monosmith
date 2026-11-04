import type { AddOptions, ChoiceOption } from './add-options.contract.ts'
import { askedAxes, choicesByAxis } from './add-options.mapper.ts'
import { askedRows, singleChoiceRows, type ChoiceRow } from './add-choice-rows.mapper.ts'
import { whatIs } from './add-outcomes.mapper.ts'

/** What a question about one axis says about itself. */
export interface AxisQuestion {
  /** The axis (`shell`, `style`, `test`). */
  axis: string
  /** What the axis is called to a person ("desktop shell"). */
  name: string
  /** One line that says what the words of the axis mean. */
  hint: string
}

/** What the questions of the choices need from the editor: one list for each way of answering. */
export interface ChoicePrompts {
  /** Several answers: the rows ticked at first are the defaults. Resolves undefined on Escape. */
  many:  (question: AxisQuestion, rows: ReadonlyArray<ChoiceRow & { picked: boolean }>) => Promise<string[] | undefined>
  /** One answer: the first row is the one Enter takes. Resolves the row, or undefined on Escape. */
  one:   (question: AxisQuestion, rows: readonly ChoiceRow[]) => Promise<ChoiceRow | undefined>
  /** One answer on an axis smith never picks (a trigger, a runtime). Resolves the name, or undefined on Escape. */
  asked: (question: AxisQuestion, rows: readonly ChoiceRow[], needing: readonly string[]) => Promise<string | undefined>
  /** Says what is wrong with the picks and asks whether to pick again. Resolves true to ask again, false to cancel. */
  again: (problem: string) => Promise<boolean>
}

/** Checks the picks with smith; resolves what is wrong or missing in plain words, or '' when they are fine. */
export type PicksCheck = (picked: readonly string[]) => Promise<string>

/**
 * Asks the choices of a new project, axis by axis. What the first questions
 * settled is kept and not asked again (a website's framework, a function's
 * cloud); a list of one answer starts on its default, or on "None" when the axis
 * has no default, so Enter alone never adds a desktop shell or a style.
 * @param options - what the questions offer
 * @param language - the language of the new project
 * @param preset - the choices the first questions named (a framework, a desktop shell, a cloud)
 * @param outcomeId - what is being built; undefined on the long way
 * @param prompts - how to ask
 * @param role - the kind of project (`app`, `lib`, `package`): the lists show only the choices that fit it
 * @param check - checks the picks with smith before the list is left; a problem is said and the questions start again
 * @returns the names to give `--with`; undefined when the person pressed Escape
 */
export async function pickChoices (options: AddOptions, language: string, preset: readonly string[], outcomeId: string | undefined, prompts: ChoicePrompts, role?: string, check?: PicksCheck): Promise<string[] | undefined> {
  for (;;) {
    const picked = await askAxes(options, language, preset, outcomeId, prompts, role)
    if (picked === undefined || check === undefined) return picked
    const problem = await check(picked)
    if (problem === '') return picked
    if (!await prompts.again(problem)) return undefined
  }
}

async function askAxes (options: AddOptions, language: string, preset: readonly string[], outcomeId: string | undefined, prompts: ChoicePrompts, role: string | undefined): Promise<string[] | undefined> {
  const axes = choicesByAxis(options.choices, language, role)
  const picked: string[] = [...preset]
  const settled = new Set(axes.filter((axis) => axis.choices.some((choice) => preset.includes(choice.name))).map((axis) => axis.axis))
  const asked = askedAxes(options.choices, language)
  const chosen = (): ChoiceOption[] => options.choices.filter((choice) => picked.includes(choice.name))
  for (const axis of axes) {
    if (settled.has(axis.axis)) continue
    const question: AxisQuestion = { axis: axis.axis, name: options.layers[axis.axis]?.toLowerCase() ?? axis.axis, hint: whatIs(axis.axis) }
    if (asked.includes(axis.axis)) {
      const needing = chosen().filter((choice) => choice.needs.includes(axis.axis))
      if (needing.length === 0) continue
      const answer = await prompts.asked(question, askedRows(axis.choices, chosen()), needing.map((choice) => choice.name))
      if (answer === undefined) return undefined
      picked.push(answer)
      continue
    }
    if (axis.many) {
      const rows = axis.choices.map((choice) => ({
        name:        choice.name,
        label:       choice.title ?? choice.name,
        description: `${choice.isDefault ? '(default) ' : ''}${choice.layer === undefined ? '' : `${options.layers[choice.layer] ?? choice.layer} `}${choice.requires.length > 0 ? `needs ${choice.requires.join(', ')}` : ''}`.trim(),
        detail:      choice.description,
        picked:      choice.isDefault,
      }))
      const answer = await prompts.many(question, rows)
      if (answer === undefined) return undefined
      picked.push(...answer)
      continue
    }
    const answer = await prompts.one(question, singleChoiceRows(axis, chosen(), outcomeId, options.layers))
    if (answer === undefined) return undefined
    if (answer.name !== '') picked.push(answer.name)
  }

  return picked
}
