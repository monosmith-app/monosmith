/** One thing `smith add <kind>` accepts, as `smith kinds --json` lists it. */
export interface KindOption {
  /** The combined kind, such as `node-app` or `go-lib`. */
  kind:        string
  language:    string
  label:       string
  description: string
}

/** One choice of `smith choices --json`: a framework, a test runner, a style. */
export interface ChoiceOption {
  axis:        string
  name:        string
  description: string
  /** The language it fits; absent when it fits every language. */
  language?:   string
  /** The choices it needs (`chakra` needs `react`). */
  requires:    string[]
  /** An axis left out takes its default. */
  isDefault:   boolean
  /** The axis allows several choices at once. */
  many:        boolean
}

/** One stack of `smith stacks --json`. */
export interface StackOption {
  name:        string
  description: string
  language:    string
}

/** Everything the questions offer. */
export interface AddOptions {
  kinds:   KindOption[]
  choices: ChoiceOption[]
  stacks:  StackOption[]
}

/** What the person asked to add. */
export interface AddRequest {
  /** `app`, `lib` or `package`. */
  role:     string
  name:     string
  language: string
  /** A stack from `smith stacks`; absent for the default of the language. */
  stack?:   string
  /** Choices to build the project from (`--with`). */
  with:     string[]
  /** Give the stack's defaults instead of following the sibling projects. */
  fresh:    boolean
}

/** What the person decided after reading the plan. */
export type PlanAnswer = 'add' | 'fresh' | 'cancel'

/** How adding a project ended. */
export type AddOutcome =
  | { kind: 'added', name: string, lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
