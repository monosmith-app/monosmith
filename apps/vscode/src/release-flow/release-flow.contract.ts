/** A package the release plan would publish, as far as the choices read it. */
export interface PlannedPackage {
  name:     string
  /** The version it moves to. */
  next:     string
  /** `major`, `minor` or `patch`. */
  level:    string
  /** The last released version; absent for a package that was never released (smith refuses a bump for it). */
  current?: string
}

/** What the person decided before the release: the packages to hold back and the bumps to override. */
export interface ReleaseChoices {
  /** Packages that do not release this time. */
  hold: string[]
  /** Overrides of the level, `<project>=patch|minor|major`. */
  bump: string[]
}

/** How the release flow ended: published (`lines` are what smith printed), nothing to release (`lines` say why), refused, or cancelled. */
export type ReleaseOutcome =
  | { kind: 'released', lines: string[] } |
  { kind: 'nothing', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
