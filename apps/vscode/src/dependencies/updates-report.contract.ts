/** Where a dependency is declared. */
export interface UpdateSite {
  file:        string
  name:        string
  spec:        string
  section:     string
  rewritable?: boolean
}

/** A declared dependency with a newer release, as `smith up --json` lists it. */
export interface DependencyUpdate {
  plugin:  string
  name:    string
  current: string
  latest:  string
  /** `patch`, `minor` or `major`. */
  level:   string
  sites?:  UpdateSite[]
}

/** A tool smith pins (Node.js, Go, ...) compared with its newest release. */
export interface ToolUpdate {
  plugin: string
  name:   string
  pinned: string
  latest: string
  newer?: boolean
}

/** What `smith up --json` prints. Unknown fields are ignored. */
export interface UpdatesReport {
  updates?: DependencyUpdate[]
  tools?:   ToolUpdate[]
}
