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

/** A folder of the editor and what smith said about it. */
export interface FolderWorkspace {
  root:     string
  name:     string
  snapshot: WorkspaceSnapshot
}

/** The workspace at one moment: its projects and which of them are affected. */
export interface WorkspaceSnapshot {
  projects: WorkspaceProject[]
  affected: ReadonlySet<string>
}
