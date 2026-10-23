/** One project, as `smith projects --json` describes it. */
export interface WorkspaceProject {
  name:     string
  dir:      string
  /** `app`, `lib` or `package`. */
  kind:     string
  language: string
  tasks:    string[]
  deps:     string[]
}

/** The workspace at one moment: its projects and which of them are affected. */
export interface WorkspaceSnapshot {
  projects: WorkspaceProject[]
  affected: ReadonlySet<string>
}
