/** What a run will do, before it has an id. */
export interface RunPlan {
  /** The task names to run. */
  tasks:     string[]
  /** The projects to run them in; all that have them when absent. */
  projects?: string[]
  /** Only the projects the changes affect. */
  affected?: boolean
  /** Stop at the first failure. */
  bail?:     boolean
  /** Run again even when the cache has the result. */
  noCache?:  boolean
}

/**
 * What the person asked to run: one task of one project; every task of some
 * projects (a project, or a whole group of them); one task in the affected
 * projects; or one task in every project that has it.
 */
export type RunTarget = { kind: 'task', project: string, task: string } |
  { kind: 'projects', projects: string[] } |
  { kind: 'affected', task: string } |
  { kind: 'all', task: string }

/** The options that apply to any run. */
export interface RunOptions {
  bail:    boolean
  noCache: boolean
}
