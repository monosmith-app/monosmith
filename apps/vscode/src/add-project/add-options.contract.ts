import type { LanguageInfo, OutcomeOption } from './add-outcomes.mapper.ts'
import type { DoctorFailure } from '../smith-run/index.ts'

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
  /** The axes it needs an answer on; an axis with no default is asked (a function host needs a trigger and a runtime). */
  needs:       string[]
  /** The answer a question pre-selects; smith never adds it by itself. */
  suggested:   boolean
  /** The choices it does not go with. */
  conflicts:   string[]
  /** What a person calls it ("Electron"); absent when the name is enough. */
  title?:      string
  /** Where it sits in the stack (`framework`, `shell`, `host`); absent means the layer of its axis. */
  layer?:      string
  /** Words that find it ("desktop", "serverless"). */
  aliases?:    string[]
  /** The kinds of project it fits (`app`, `lib`, `package`); absent means every kind, as for a lint tool or a test runner. */
  kinds?:      string[]
}

/** One stack of `smith stacks --json`. */
export interface StackOption {
  name:        string
  description: string
  language:    string
}

/** Everything the questions offer. */
export interface AddOptions {
  kinds:     KindOption[]
  choices:   ChoiceOption[]
  stacks:    StackOption[]
  /** What a person can build, as `smith compose --json` lists it (the list of the terminal screen). */
  outcomes:  OutcomeOption[]
  /** The languages with the names a person knows them by. */
  languages: LanguageInfo[]
  /** The title of each layer of the catalogue (`shell` is "Desktop shell"): the one name for a framework, a shell or a host. */
  layers:    Record<string, string>
}

/** What the person asked to add. */
export interface AddRequest {
  /** `app`, `lib` or `package`. */
  role:          string
  name:          string
  language:      string
  /** A stack from `smith stacks`; absent for the default of the language. */
  stack?:        string
  /** Choices to build the project from (`--with`). */
  with:          string[]
  /** Give the stack's defaults instead of following the sibling projects. */
  fresh:         boolean
  /** Copy this one project's shape instead of the majority's (`--like`). */
  like?:         string
  /** The environment builds of a web app (`--environments` and the flags with it); absent for a project that has one build. */
  environments?: WebEnvironments
}

/** How a web app is built once per environment: what `smith add` asks and the flags answer. */
export interface WebEnvironments {
  names:    string[]
  /** The environment the task `build` builds, which CI and releases use. */
  default:  string
  /** `workspace`: dist/apps/<app>/<env>; `project`: dist/<env> in the app. */
  output:   'workspace' | 'project'
  /** `commit`: git keeps .env.<env>; `example`: git keeps only .env.example. */
  envFiles: 'commit' | 'example'
}

/**
 * What the person decided after reading the plan. The -commit answers tick
 * the box that is unticked by default: make the change one commit, so Undo works.
 */
export type PlanAnswer = 'add' | 'add-commit' | 'fresh' | 'fresh-commit' | 'cancel'

/** A tool that stopped the first check of a new project (`smith add` made the project and kept it). */
export interface MissingTool {
  /** What the person calls it: "the .NET SDK". */
  title:  string
  /** The page that tells how to get it; absent when smith printed none. */
  url?:   string
  /** Where the project is; absent when smith printed none. */
  dir?:   string
  /** How many files smith wrote; 0 when unknown. */
  files:  number
  /** The smith arguments that run the first check again; absent when smith printed none. */
  check?: string[]
}

/** The tools a new project needs that the machine lacks or has in another version, as `smith add` reports them. */
export interface NeededTools {
  /** The names `smith tools install` takes: `golangci-lint`, `python`. */
  names:    string[]
  /** What is wrong with each, in plain words ("golangci-lint is 2.13.0, not the 2.15.0 that CI pins: ..."). */
  problems: string[]
}

/**
 * How adding a project ended. `added` carries `tools` when the project needs
 * tools the machine lacks: smith installs nothing without a terminal, so the
 * editor offers the choice (`smith tools install`). `unchecked` is a success
 * that could not be checked: the project is on disk, a tool it needs is
 * missing. `failing` is a project that is on disk and fails the checks of
 * `smith doctor`.
 */
export type AddOutcome =
  | { kind: 'added', name: string, role: string, operation: string, lines: string[], committed: boolean, tools?: NeededTools } |
  { kind: 'unchecked', name: string, tool: MissingTool, lines: string[] } |
  { kind: 'failing', name: string, operation: string, failure: DoctorFailure, lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'composing' } |
  { kind: 'cancelled' }
