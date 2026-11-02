/** One answer a person can pick. */
export interface GuidedOption {
  label:        string
  value:        string
  description?: string
}

/** Where the options or the first value of a step come from, when smith holds them and not this extension. */
export type GuidedSource =
  | 'web-apps' | 'environments-of-app' | 'workspace-environments' | 'pipeline-templates' |
  'settings' | 'settings-set' | 'setting-value' | 'builtin-plugins' | 'architectures' | 'marketplace'

/** One question of a guided command. */
export interface GuidedStep {
  key:          string
  /** `text` is an input box, `pick` a list of one answer, `multi` a list with tick boxes, `file` the file dialog. */
  kind:         'text' | 'pick' | 'multi' | 'file'
  title:        string
  prompt?:      string
  placeholder?: string
  /** The text an input box starts on, which the person edits. */
  value?:       string
  options?:     GuidedOption[]
  /** Reads the options (or, for `text`, the first value) from smith. */
  from?:        GuidedSource
  /** An answer may be left empty. */
  optional?:    boolean
  /** Asks the step only when this says so, given the earlier answers. */
  when?:        (answers: GuidedAnswers) => boolean
  /** Says what is wrong with a typed answer, or undefined. */
  validate?:    (value: string) => string | undefined
}

/** What the person answered, by step key. */
export type GuidedAnswers = Record<string, string | string[]>

/** A command of smith that has a screen of its own in the extension. */
export interface GuidedCommand {
  /** The suffix of the VS Code command: `monosmith.<id>`. */
  id:           string
  title:        string
  /** `read` shows the output; `change` shows the plan or the consequence, asks, and then runs. */
  mode:         'read' | 'change'
  steps:        GuidedStep[]
  /** The smith arguments the answers build. */
  args:         (answers: GuidedAnswers) => string[]
  /** The flag that makes smith print its plan and change nothing (`--dry-run`). */
  planFlag?:    string
  /** True when the real run needs `--yes` (the command asks on a terminal). */
  yes?:         boolean
  /** The button that confirms, a short verb. */
  verb?:        string
  /** What the change does and what it costs, said before the question when there is no plan. */
  consequence?: string
  /** The command takes `--commit`: the unticked "Commit this change so it can be undone" box comes before the confirmation. */
  commit?:      boolean
}

/** A step with its options and first value read, ready to be asked. */
export interface ResolvedStep extends GuidedStep {
  options: GuidedOption[]
}

/** How a guided command ended. */
export type GuidedOutcome =
  | { kind: 'shown' } |
  { kind: 'done', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
