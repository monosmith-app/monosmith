/** One question of a fix: the person picks one answer. The first is the suggested one. */
export interface DoctorFixChoice {
  question: string
  options:  Array<{ label: string }>
}

/** The change that clears a failed check, as `smith doctor --json` carries it. */
export interface DoctorFix {
  description: string
  /** True when `smith doctor --fix --yes` alone may apply it. */
  safe?:       boolean
  choices?:    DoctorFixChoice[]
  /** The files the fix writes; the first one is the file a Health row names. */
  files?:      Array<{ path: string }>
  /** The projects a fix folded from many projects applies to. */
  applies_to?: string[]
}

/** One check of `smith doctor --json`. */
export interface DoctorFinding {
  check:    string
  ok:       boolean
  /** What is wrong, in one sentence. */
  problem?: string
  /** The project the check is about; absent for the workspace. */
  project?: string
  detail?:  string
  remedy?:  string
  /** A failed check that only warns. */
  warning?: boolean
  /** The fix smith can apply after the person agrees. */
  fix?:     DoctorFix
}

/** What `smith doctor --json` prints. Unknown fields are ignored. */
export interface DoctorReport {
  findings?: DoctorFinding[]
}
