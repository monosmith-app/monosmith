import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import { parseNeverReleasedPackages, parseUnchangedPackages } from './forced-release.mapper.ts'
import { parsePlannedPackages, releaseArguments } from './release-arguments.mapper.ts'
import type { PlannedPackage, ReleaseChoices, ReleaseOutcome } from './release-flow.contract.ts'

/** What releasing needs from the editor. */
export interface ReleaseDeps {
  /** Runs smith for a short answer (the plans). */
  run:           (args: string[]) => Promise<SmithRun>
  /** Runs smith to its end (the release), with progress shown. */
  apply:         (args: string[]) => Promise<SmithRun>
  /** Shows the plan and asks which packages to hold back; undefined means the person cancelled. */
  chooseHold:    (plan: string, packages: PlannedPackage[]) => Promise<string[] | undefined>
  /** Asks which bumps to override, among the packages not held; undefined means the person cancelled. */
  chooseBump:    (packages: PlannedPackage[]) => Promise<string[] | undefined>
  /**
   * Offers to release a package no commit releases (the same code under a new version), with its
   * consequences; undefined means the person cancelled. Without it the path is not offered.
   */
  chooseForced?: (unchanged: string[], nothingPlanned: boolean, neverReleased: string[]) => Promise<string[] | undefined>
  /** Shows the final plan and states that a release cannot be undone; false means the person cancelled. */
  confirm:       (plan: string, choices: ReleaseChoices) => Promise<boolean>
}

/**
 * Releases the workspace with the choices of the person. It plans (`--dry-run`), asks which
 * packages to hold back and which bumps to override, plans again when a choice was made,
 * and only after the final question runs `smith release --yes`. A release tags and publishes,
 * and Undo does not revert it.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function releaseWorkspace (deps: ReleaseDeps): Promise<ReleaseOutcome> {
  const text = await deps.run(['release', '--dry-run'])
  if (text.code !== 0) return { kind: 'refused', message: describeFailure(text) }
  const json = await deps.run(['release', '--dry-run', '--json'])
  if (json.code !== 0) return { kind: 'refused', message: describeFailure(json) }
  let packages: PlannedPackage[]
  try {
    packages = parsePlannedPackages(json.stdout)
  } catch {
    return { kind: 'refused', message: 'smith printed the release plan in a form this editor does not read. Update the extension, or run smith release in a terminal.' }
  }
  let unchanged: string[]
  let never: string[]
  try {
    unchanged = parseUnchangedPackages(json.stdout)
    never = parseNeverReleasedPackages(json.stdout)
  } catch {
    unchanged = []
    never = []
  }
  // smith refuses a forced bump for a package that was never released: it is told apart, not offered.
  unchanged = unchanged.filter((name) => !never.includes(name))
  let forced: string[] = []
  if (deps.chooseForced !== undefined && (unchanged.length > 0 || never.length > 0)) {
    const asked = await deps.chooseForced(unchanged, packages.length === 0, never)
    if (asked === undefined) return { kind: 'cancelled' }
    forced = asked
  }
  if (packages.length === 0 && forced.length === 0) return { kind: 'nothing', lines: text.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean), unchanged }

  const hold = packages.length === 0 ? [] : await deps.chooseHold(text.stdout.trim(), packages)
  if (hold === undefined) return { kind: 'cancelled' }
  const bump = packages.length === 0 ? [] : await deps.chooseBump(packages.filter((item) => !hold.includes(item.name)))
  if (bump === undefined) return { kind: 'cancelled' }
  const choices: ReleaseChoices = { hold, bump, ...(forced.length > 0 && { forced }) }

  const final = hold.length === 0 && bump.length === 0 && forced.length === 0 ? text : await deps.run(releaseArguments(choices, 'plan'))
  if (final.code !== 0) return { kind: 'refused', message: describeFailure(final) }
  if (!await deps.confirm(final.stdout.trim(), choices)) return { kind: 'cancelled' }

  const done = await deps.apply(releaseArguments(choices, 'publish'))

  return done.code === 0 ? { kind: 'released', lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean) } : { kind: 'refused', message: describeFailure(done) }
}
