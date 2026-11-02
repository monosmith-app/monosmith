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

/** One answer of a question, in the words a person knows. */
export interface ChoiceLabel {
  value:       string
  label:       string
  description: string
}

/** How creating a workspace ended. `terminal` means the person asked for smith's own questions in a terminal. */
export type NewWorkspaceOutcome =
  | { kind: 'created', name: string, lines: string[], committed: boolean } |
  { kind: 'terminal' } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
