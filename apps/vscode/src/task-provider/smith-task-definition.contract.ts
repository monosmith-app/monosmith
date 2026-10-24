/**
 * A `smith` task of tasks.json: what `smith run` is asked for.
 * ```json
 * { "type": "smith", "task": "test", "project": "web", "affected": true }
 * ```
 */
export interface SmithTaskDefinition {
  type:      'smith'
  /** The task to run, such as `build`. */
  task:      string
  /** The projects to run it in (names, globs, `kind:app`); every project that has it when absent. */
  project?:  string
  /** Only the projects the changes affect. */
  affected?: boolean
  bail?:     boolean
  noCache?:  boolean
  /** How many tasks may run at once. */
  parallel?: number
}

/** A task the provider offers without being asked: Run Task lists it. */
export interface OfferedTask {
  /** The name Run Task shows after `smith:`. */
  name:       string
  definition: SmithTaskDefinition
  /** `build` and `test` tasks join the default build and test task groups. */
  group?:     'build' | 'test' | 'clean'
}
