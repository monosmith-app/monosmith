/**
 * Reads the packages that no commit releases from `smith release --dry-run --json`.
 * @param json - what smith printed
 * @returns the names of the unchanged packages, in smith's order
 * @throws SyntaxError when the text is not JSON
 */
export function parseUnchangedPackages (json: string): string[] {
  const root: unknown = JSON.parse(json)
  const unchanged = typeof root === 'object' && root !== null ? (root as { unchanged?: unknown }).unchanged : undefined

  return Array.isArray(unchanged) ? unchanged.filter((name): name is string => typeof name === 'string' && name !== '') : []
}

/**
 * Reads the unchanged packages that were never released from `smith release --dry-run --json`.
 * smith refuses to force a bump for them, because there is no version to bump.
 * @param json - what smith printed
 * @returns the names; none when smith does not say
 * @throws SyntaxError when the text is not JSON
 */
export function parseNeverReleasedPackages (json: string): string[] {
  const root: unknown = JSON.parse(json)
  const never = typeof root === 'object' && root !== null ? (root as { neverReleased?: unknown }).neverReleased : undefined

  return Array.isArray(never) ? never.filter((name): name is string => typeof name === 'string' && name !== '') : []
}

/**
 * The commit that sets the first version of a package that was never released.
 * @param name - the package
 */
export function firstVersionCommand (name: string): string {
  return `git commit --allow-empty -m "version(${name})[1.0.0]"`
}

/**
 * Says in plain words what a release without a releasing commit does. The final question shows it.
 * @param forced - `<project>=patch|minor|major` for each package released this way
 * @returns the sentence, or an empty text when no package is released this way
 */
export function forcedConsequences (forced: string[]): string {
  if (forced.length === 0) return ''
  const names = forced.map((item) => item.split('=', 1)[0]).join(', ')

  return `\n\n${names}: no commit releases it. The new tag points at HEAD, and the same code is published under a new version, with a GitHub Release and the assets.`
}
