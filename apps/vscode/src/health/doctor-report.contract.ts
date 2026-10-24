/** One check of `smith doctor --json`. */
export interface DoctorFinding {
  check:    string
  ok:       boolean
  detail?:  string
  remedy?:  string
  /** A failed check that only warns. */
  warning?: boolean
}

/** What `smith doctor --json` prints. Unknown fields are ignored. */
export interface DoctorReport {
  findings?: DoctorFinding[]
}
