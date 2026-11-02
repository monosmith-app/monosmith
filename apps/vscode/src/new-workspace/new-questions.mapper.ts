import type { NewAnswer, NewQuestion } from './new-workspace.contract.ts'

function text (value: unknown): string {
  return typeof value === 'string' ? value : ''
}

/**
 * Reads what `smith new --questions --json` printed.
 * @param json - the text smith printed
 * @returns the questions in smith's order
 * @throws Error when the text is not a JSON list
 */
export function parseNewQuestions (json: string): NewQuestion[] {
  const list: unknown = JSON.parse(json)
  if (!Array.isArray(list)) throw new TypeError('smith did not print a list of questions')

  return list
    .filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null && text((item as Record<string, unknown>).key) !== '')
    .map((item) => ({
      key:      text(item.key),
      prompt:   text(item.prompt),
      default:  text(item.default),
      flag:     text(item.flag),
      many:     item.many === true,
      advanced: item.advanced === true,
      atEnd:    item.atEnd === true,
      when:     Object.fromEntries(Object.entries((item.when ?? {}) as Record<string, unknown>).map(([key, value]) => [key, text(value)])),
      answers:  (Array.isArray(item.options) ? item.options as Array<Record<string, unknown>> : [])
        .map((option): NewAnswer => ({ value: text(option.value), label: text(option.label), hint: text(option.hint) })),
    }))
}

/**
 * The questions the extension asks: not advanced, not asked at the end, with a flag and a list
 * of answers.
 * @param questions - every question of `smith new`
 * @returns the questions to draw, in order
 */
export function askableQuestions (questions: readonly NewQuestion[]): NewQuestion[] {
  return questions.filter((question) => !question.advanced && !question.atEnd && question.flag !== '' && question.answers.length > 0)
}

/**
 * Whether a question applies for the answers given so far.
 * @param question - the question
 * @param answers - the answers so far, key to value
 * @returns true when every answer its `when` names holds
 */
export function appliesWhen (question: NewQuestion, answers: Readonly<Record<string, string>>): boolean {
  return Object.entries(question.when).every(([key, value]) => answers[key] === value)
}
