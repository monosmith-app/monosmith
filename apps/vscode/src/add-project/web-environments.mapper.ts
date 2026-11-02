import type { AddRequest, WebEnvironments } from './add-options.contract.ts'

/** The environments a web app starts from: lcl is the developer's machine. */
export const DEFAULT_ENVIRONMENTS = ['lcl', 'dev', 'uat', 'prd']

/** Names Vite or the env files already give a meaning to. */
const RESERVED = new Set(['development', 'production', 'test', 'local', 'example'])

/**
 * Whether the request makes a web app built with Vite (react, vue, svelte or a plain page),
 * which is built once per environment.
 * @param request - what the person chose
 */
export function isWebApp (request: Pick<AddRequest, 'role' | 'language' | 'with' | 'stack'>): boolean {
  if (request.role !== 'app' || request.language !== 'node') return false

  return [...request.with, ...(request.stack === undefined ? [] : [request.stack])].some((name) => /^(?:react|vue|svelte|html)(?:-|$)/.test(name))
}

/**
 * The names a person typed, as a list: separated by commas, spaces or semicolons.
 * @param text - what the person typed
 */
export function parseEnvironmentNames (text: string): string[] {
  return text.split(/[\s,;]+/).filter((name) => name !== '')
}

/**
 * What is wrong with the environment names a person typed, in words that name the fix.
 * @param text - what the person typed
 * @returns the problem, or undefined when the names are fine; smith has the last word
 */
export function checkEnvironmentNames (text: string): string | undefined {
  const names = parseEnvironmentNames(text)
  if (names.length === 0) return 'Name at least one environment, for example lcl, dev, uat, prd.'
  for (const name of names) {
    if (!/^[a-z][a-z0-9-]*$/.test(name)) return `"${name}" is not an environment name: use a lowercase word, for example uat.`
    if (RESERVED.has(name)) return `"${name}" cannot be an environment name: Vite or the env files already give it a meaning. Use another word, for example lcl for local.`
  }

  return new Set(names).size === names.length ? undefined : 'An environment is listed twice.'
}

/**
 * The environment that `smith run build` builds when the person does not pick one: prd, else the last.
 * @param names - the environments
 */
export function defaultEnvironment (names: readonly string[]): string {
  return names.includes('prd') ? 'prd' : (names.at(-1) ?? 'prd')
}

/**
 * The flags of `smith add` for the environments of a web app.
 * @param environments - what the person decided
 */
export function environmentArguments (environments: WebEnvironments): string[] {
  return [
    '--environments', environments.names.join(','),
    '--default-environment', environments.default,
    '--environment-output', environments.output,
    '--env-files', environments.envFiles,
  ]
}
