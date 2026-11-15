/** A failure found in the output of a task. */
export interface ParsedDiagnostic {
  /** The path as the tool printed it: relative to where the task ran, or absolute. */
  file:     string
  /** 1-based. */
  line:     number
  /** 1-based; 1 when the tool printed none. */
  column:   number
  severity: 'error' | 'warning' | 'info'
  message:  string
  /** The tool's code for it, such as `TS2322` or an ESLint rule. */
  code?:    string
}

/** A reference to a place in a file inside a line of output. */
export interface PathReference {
  /** Where it starts in the line (0-based) and where it ends (exclusive). */
  start:   number
  end:     number
  file:    string
  line:    number
  column?: number
}
