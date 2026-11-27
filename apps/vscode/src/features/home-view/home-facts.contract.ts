/** The last run of the session, in a few facts. */
export interface HomeRun {
  /** The smith command line it stands for. */
  command: string
  /** The first task that failed (`orders:test`), when one did. */
  task?:   string
  running: boolean
  ok:      boolean
  stopped: boolean
  /** How long it took, already formatted. */
  took:    string
  /** How long ago it ended, `3 min ago`. */
  ago:     string
  passed:  number
  cached:  number
  failed:  number
}

/** What doctor found, in a few facts. */
export interface HomeHealth {
  failed:       number
  warnings:     number
  passed:       number
  /** How many failed checks smith can fix. */
  fixable:      number
  /** The failed checks, by name, the first one first. */
  problems:     string[]
  /** The checks that warn, by name, the first one first. */
  warningNotes: string[]
}

/** The next release in a few facts: how many packages leave, and the one the sentence names. */
export interface HomeRelease {
  packages: number
  name:     string
  /** The version now; empty before the first release. */
  from:     string
  to:       string
  /** How many commits the named package releases (0 when not known). */
  commits:  number
  major:    number
  minor:    number
  patch:    number
}

/** What Home is drawn from: what the extension already holds, so drawing never waits. */
export interface HomeFacts {
  /** True once the projects of every folder arrived. */
  projectsLoaded: boolean
  projects:       number
  /** The language of each project. */
  languages:      string[]
  /** How many projects the changes affect. */
  affected:       number
  /** What doctor found; undefined while it has not answered. */
  health?:        HomeHealth
  /** Why doctor could not run. */
  healthError?:   string
  /** The newest run of the session; undefined before the first. */
  last?:          HomeRun
  /** The next release; undefined while unknown or when nothing is ready. */
  release?:       HomeRelease
  /** True once the release plan was read. */
  releaseKnown:   boolean
  /** How many dependencies have a newer release; undefined while the registries were not asked. */
  outdated?:      number
}

/** How much an item of the queue asks for: it sets the mark and the colour. */
export type HomeSeverity = 'failure' | 'warning' | 'info' | 'muted' | 'success'

/** A button of an item: what it says, the command it runs, and the smith command line it stands for. */
export interface HomeButton {
  label:   string
  command: string
  /** Absent for a button that only opens a view. */
  line?:   string
}

/** One item of "Needs you now". */
export interface HomeItem {
  /** What kind of item it is, the tag the terminal shows (`Failed`, `Warning`, `Changed`...). */
  kind:     'empty' | 'checking' | 'doctor' | 'run' | 'setup' | 'warnings' | 'changes' | 'release' | 'updates' | 'good'
  severity: HomeSeverity
  title:    string
  detail?:  string
  /** The first one is what choosing the item runs. */
  buttons:  HomeButton[]
}
