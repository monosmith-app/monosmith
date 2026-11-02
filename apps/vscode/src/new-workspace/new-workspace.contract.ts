/** What the person decides about a new workspace: the questions of `smith new`, one field each. */
export interface NewWorkspaceChoices {
  /** The workspace name, in kebab-case. */
  name:         string
  /** Where the tests run on every push: `github`, `azure`, `gitlab`, `circleci` or `bitbucket`. */
  ci:           string[]
  /** The main language, or `later`. */
  language:     string
  /** The tests and lint of Node projects (`node-jest`, `node-vitest`, `node-ts`); used only for `node`. */
  nodeStack:    string
  /** The code layout smith checks, or `none`. */
  architecture: string
  /** Write the MonoSmith agent files. */
  agent:        boolean
  /** Write the dev container. */
  devcontainer: boolean
  /** Make the first commit of smith's own files. */
  commit:       boolean
}

/** One answer of a question, as `smith new --questions --json` prints it. */
export interface NewAnswer {
  value: string
  label: string
  /** What the answer means, or empty. */
  hint:  string
}

/**
 * One question of `smith new`, as `smith new --questions --json` prints it: the data the CLI and the
 * TUI ask from.
 */
export interface NewQuestion {
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
  answers:  NewAnswer[]
}

/** How creating a workspace ended. `terminal` means the person asked for smith's own questions in a terminal. */
export type NewWorkspaceOutcome =
  | { kind: 'created', name: string, lines: string[], committed: boolean } |
  { kind: 'terminal' } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
