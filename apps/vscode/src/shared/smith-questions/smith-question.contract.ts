/** One answer of a question, as `smith new --questions --json` prints it. */
export interface SmithAnswer {
  value: string
  label: string
  /** What the answer means, or empty. */
  hint:  string
}

/**
 * One question of `smith new`, as `smith new --questions --json` prints it: the data the CLI and the
 * TUI ask from.
 */
export interface SmithQuestion {
  key:      string
  prompt:   string
  default:  string
  /** The flag that answers it; empty for a question that only opens more questions. */
  flag:     string
  /** The answer may name several options. */
  many:     boolean
  /** Asked only after "more options": the extension does not ask it. */
  advanced: boolean
  /** Asked after the summary: the extension does not ask it. */
  atEnd:    boolean
  /** The answers (key to value) that make this question apply. */
  when:     Record<string, string>
  /** The answers on offer; none means free text. */
  answers:  SmithAnswer[]
}

/**
 * Shows one question and returns the values picked: one for a single choice, several for a question
 * that takes many. Resolves undefined when the person cancels.
 */
export type PickAnswer = (question: SmithQuestion, current: readonly string[]) => Promise<string[] | undefined>
