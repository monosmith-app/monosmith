/** One package a release would publish, as `smith release --dry-run --json` plans it. */
export interface ReleaseResult {
  project:  string
  plugin?:  string
  /** The last released version; absent when never released. */
  current?: string
  next:     string
  /** `major`, `minor` or `patch`. */
  level?:   string
  tag?:     string
  /** Why the package releases. */
  reason?:  string
  commits?: string[]
  forced?:  boolean
  status?:  string
  error?:   string
}

/** What `smith release --dry-run --json` prints. Unknown fields are ignored. */
export interface ReleasePreview {
  dryRun?:    boolean
  /** The release channel of the branch (`next`, `beta`); absent on the stable line. Its versions are prereleases. */
  channel?:   string
  results?:   ReleaseResult[]
  /** The packages with nothing to release. */
  unchanged?: string[]
  /** Set when the plan stopped early. */
  error?:     string
}
