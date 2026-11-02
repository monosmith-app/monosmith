/** One thing the catalogue of `smith compose --json` lists. */
export interface ComposeItem {
  /** Unique: language/name. */
  id:          string
  /** `language`, `choice` or `plugin`. */
  kind:        string
  name:        string
  title:       string
  description: string
  aliases:     string[]
  layer:       string
  language:    string
  suggested:   boolean
  isDefault:   boolean
  /** Why the item cannot be used here; empty when it can. */
  problem:     string
  /** The command that turns a workspace plugin on; empty for an item of `--with`. */
  route:       string
}

/** One layer of the stack, from the language up to the pipelines. */
export interface ComposeLayer {
  id:    string
  title: string
  about: string
}

/** One answer to a question smith does not answer for the person. */
export interface ComposeOption {
  name:        string
  title:       string
  description: string
  suggested:   boolean
}

/** A pick needs an answer on an axis (a function host needs a trigger). */
export interface ComposeQuestion {
  choice:  string
  about:   string
  axis:    string
  options: ComposeOption[]
}

/** What the picks add up to, as smith checked them. */
export interface ComposeCheck {
  language:  string
  complete:  boolean
  /** The picks cannot go together: one sentence that names both items and the fix. */
  conflict:  string
  /** Not wrong, but not finished: the next thing to add. */
  missing:   string
  question?: ComposeQuestion
  /** The picks with the defaults added; set when complete. */
  selection: string[]
  /** The names smith added because an axis was needed and had a default. */
  added:     string[]
}

/** The plan a complete selection ends in. */
export interface ComposePlan {
  /** The exact `smith add` line. */
  command:    string
  argv:       string[]
  files:      Array<{ action: string, path: string, note: string }>
  tools:      Array<{ choice: string, name: string, binary: string, install: string }>
  notes:      string[]
  recipeArgv: string[]
  recipeNote: string
}

/** What `smith compose --json` prints. */
export interface ComposeReport {
  layers: ComposeLayer[]
  items:  ComposeItem[]
  check?: ComposeCheck
  plan?:  ComposePlan
}

/** What the person does with the plan. */
export type ComposeAnswer = 'run' | 'run-commit' | 'copy' | 'recipe' | 'cancel'

/** How composing ended. */
export type ComposeOutcome =
  | { kind: 'cancelled' } |
  { kind: 'refused', message: string } |
  { kind: 'added', name: string, lines: string[], committed: boolean } |
  { kind: 'copied', command: string } |
  { kind: 'recipe', lines: string[] }
