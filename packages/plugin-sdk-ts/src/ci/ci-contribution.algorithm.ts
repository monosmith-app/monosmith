import type { ActionStep, Credential, Step } from '../protocol/index.ts'

/**
 * A provider-native step: a GitHub action (`'actions/setup-node@v6'`) or an
 * Azure task (`'NodeTool@0'`) with its inputs, written in the order given. The
 * values of a pipeline are written as quoted YAML strings, except `true` and
 * `false`, so pass strings.
 */
export function actionStep (uses: string, inputs: Record<string, string> | Array<[string, string]> = {}): ActionStep {
  const pairs = Array.isArray(inputs) ? inputs : Object.entries(inputs)

  return pairs.length === 0 ? { uses } : { uses, inputs: pairs.map(([name, value]) => ({ name, value })) }
}

export interface StepOptions {
  /** The workspace-relative directory to run in. Default: the workspace root. */
  dir?:        string
  /** Whether a failure fails the phase (the default) or only warns. */
  blocking?:   boolean
  /** Workspace-relative files whose content decides what an audit step finds (a lockfile): smith skips a passing step while they are unchanged. */
  cacheFiles?: string[]
}

/** One command a CI phase (`setup`, `audit`, `postRelease`) runs. It blocks unless `blocking: false`. */
export function ciStep (name: string, command: string[], options: StepOptions = {}): Step {
  const { dir, cacheFiles } = options

  return { name, command, blocking: options.blocking ?? true, ...(dir !== undefined && { dir }), ...(cacheFiles !== undefined && { cacheFiles }) }
}

export interface CredentialOptions {
  /** Release works without it. Default: it is required. */
  required?: boolean
  /** Also give it to `smith ci setup`: installs from a private registry need it. */
  setup?:    boolean
  /** Also give it to every step that runs tasks. */
  tasks?:    boolean
}

/** A secret the release phase needs, read from the environment variable `env`. */
export function credential (env: string, description: string, options: CredentialOptions = {}): Credential {
  return { env, description, required: options.required ?? true, ...((options.setup === true) && { setup: true }), ...((options.tasks === true) && { tasks: true }) }
}
