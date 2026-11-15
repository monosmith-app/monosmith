import type { WebAppChoice, WebEnvironments } from './web-environments.contract.ts'
import { say } from '../words/index.ts'

/** The environments a web app starts from: lcl is the developer's machine. */
export const DEFAULT_ENVIRONMENTS = ['lcl', 'dev', 'uat', 'prd']

/** Names Vite or the env files already give a meaning to. */
const RESERVED = new Set(['development', 'production', 'test', 'local', 'example'])

/**
 * Whether the request makes a web app built with Vite (react, vue, svelte or a plain page),
 * which is built once per environment.
 * @param request - what the person chose
 */
export function isWebApp (request: WebAppChoice): boolean {
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
  if (names.length === 0) return say('vscode.web-environments.name-at-least-one-environment-for')
  for (const name of names) {
    if (!/^[a-z][a-z0-9-]*$/.test(name)) return say('vscode.web-environments.is-not-an-environment-name-use', { name })
    if (RESERVED.has(name)) return say('vscode.web-environments.cannot-be-an-environment-name-vite', { name })
  }

  return new Set(names).size === names.length ? undefined : say('vscode.web-environments.an-environment-is-listed-twice')
}

/**
 * The environment that `smith run build` builds when the person does not pick one: prd, else the last.
 * @param names - the environments
 */
export function defaultEnvironment (names: readonly string[]): string {
  return names.includes('prd') ? 'prd' : (names.at(-1) ?? 'prd')
}

/**
 * The names with the default first, so Enter in a pick list takes the default and never the first
 * name by accident (lcl is the developer's machine, not what CI or a release should build).
 * @param names - the environments
 * @param fallback - the default environment
 */
export function defaultFirst (names: readonly string[], fallback: string): string[] {
  return [fallback, ...names.filter((name) => name !== fallback)]
}

/**
 * The environments a web app starts from: the workspace's own list when it has one, else the four.
 * Every choice is the default of smith: prd (else the last name) is built, the builds go to the
 * workspace, and the .env files are kept by git.
 * @param workspaceNames - `[workspace] environments`, or empty
 */
export function startingEnvironments (workspaceNames: readonly string[]): WebEnvironments {
  const names = workspaceNames.length > 0 ? [...workspaceNames] : [...DEFAULT_ENVIRONMENTS]

  return { names, default: defaultEnvironment(names), output: 'workspace', envFiles: 'commit' }
}

/**
 * One sentence that says every environment choice, as the single question shows it.
 * @param environments - the choices
 */
export function describeEnvironments (environments: WebEnvironments): string {
  const output = environments.output === 'workspace' ? 'dist/apps/<name>/<env>' : say('vscode.web-environments.dist-env-inside-the-app')
  const files = say(environments.envFiles === 'commit' ? 'vscode.web-environments.created-and-kept-by-git' : 'vscode.web-environments.created-and-ignored-by-git-only')

  return say('vscode.web-environments.environments-smith-run-build-builds-builds', { join: environments.names.join(', '), default: environments.default, output, files })
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
