import type { AddRequest } from './add-options.contract.ts'

/**
 * The arguments of `smith add` for a request.
 * @param request - what the person chose
 * @param mode - `plan`: `--dry-run`, smith changes nothing; `apply`: `--yes`, smith writes without a question (the person already agreed)
 * @param fresh - overrides the request: give the stack's defaults and do not follow the siblings
 */
export function addArguments (request: AddRequest, mode: 'plan' | 'apply', fresh = request.fresh): string[] {
  return [
    'add', request.role, request.name, '--lang', request.language,
    ...(request.stack === undefined ? [] : ['--stack', request.stack]),
    ...(request.with.length === 0 ? [] : ['--with', request.with.join(',')]),
    ...(fresh ? ['--fresh'] : []),
    mode === 'plan' ? '--dry-run' : '--yes',
  ]
}

/**
 * Whether the plan says the project follows its siblings (smith words it
 * "follows <sibling>" in a plan and "The N siblings agree on" in a run), so
 * `--fresh` would change something.
 * @param plan - what `smith add --dry-run` printed
 */
export function followsSiblings (plan: string): boolean {
  return /\bfollows\b|\bagree on\b/i.test(plan)
}

/**
 * A name smith accepts for a project, as far as the editor can tell: kebab-case. smith has the last word.
 * @param name - what the person typed
 * @returns the problem with its fix, or undefined when the name is fine
 */
export function checkProjectName (name: string): string | undefined {
  return /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(name) ? undefined : 'Use lowercase letters, digits and single dashes, for example "web" or "billing-api".'
}
