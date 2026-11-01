/** What a run of the smith executable left behind. */
export interface SmithRun {
  /** The exit code; 1 when smith could not be started. */
  code:   number
  stdout: string
  stderr: string
}
