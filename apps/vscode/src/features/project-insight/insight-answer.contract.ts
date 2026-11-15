/** One field of a task and where its value came from. */
export interface ExplainField {
  name:   string
  value:  string
  source: string
}

/** One task of a project, as `smith explain <project> --json` resolves it. */
export interface ExplainedTask {
  id:           string
  description?: string
  fields?:      ExplainField[]
}

/** One project of `smith graph --format json`. */
export interface GraphProject {
  name:      string
  kind?:     string
  language?: string
  deps?:     string[]
  /** The project that was asked for. */
  marked?:   boolean
}

/** What `smith graph --format json` prints. Unknown fields are ignored. */
export interface GraphAnswer {
  projects?: GraphProject[]
}
