import type { PlannedPackage, ReleaseChoices } from './release-flow.contract.ts'

/**
 * Reads the packages of `smith release --dry-run --json`.
 * @param json - what smith printed
 * @returns the packages the release would publish, in smith's order
 * @throws SyntaxError when the text is not JSON
 */
export function parsePlannedPackages (json: string): PlannedPackage[] {
  const root: unknown = JSON.parse(json)
  const results = typeof root === 'object' && root !== null && Array.isArray((root as { results?: unknown }).results) ? (root as { results: unknown[] }).results : []

  return results
    .map((item) => item as { project?: unknown, next?: unknown, level?: unknown, current?: unknown })
    .filter((result): result is typeof result & { project: string } => typeof result.project === 'string' && result.project !== '')
    .map((result) => ({
      name:  result.project,
      next:  typeof result.next === 'string' ? result.next : '',
      level: typeof result.level === 'string' ? result.level : '',
      ...(typeof result.current === 'string' && result.current !== '' && { current: result.current }),
    }))
}

/**
 * The arguments of `smith release` for the choices.
 * @param choices - the packages to hold and the bumps to override
 * @param mode - `plan` adds `--dry-run`; `publish` adds `--yes`
 * @returns the arguments, without the program name
 */
export function releaseArguments (choices: ReleaseChoices, mode: 'plan' | 'publish'): string[] {
  return ['release', mode === 'plan' ? '--dry-run' : '--yes', ...(choices.hold.length === 0 ? [] : ['--hold', choices.hold.join(',')]), ...choices.bump.flatMap((bump) => ['--bump', bump]), ...(choices.forced ?? []).flatMap((bump) => ['--bump', bump]), ...((choices.forced ?? []).length === 0 ? [] : ['--force'])]
}
