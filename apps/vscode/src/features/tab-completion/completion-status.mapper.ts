/** What `smith completion status --json` prints. */
export interface CompletionStatus {
  /** The shell smith found, empty when none. */
  shell:  string
  /** Nothing is installed for it, no package ships the file, and the person did not say no. */
  wanted: boolean
  /** The answer to the offer on the command line: `no`, `later`, `installed` or empty. */
  answer: string
}

/**
 * Reads the output of `smith completion status --json`.
 * @param stdout - what smith printed
 * @returns the status, or undefined when the text is not that JSON (an older smith has no such command)
 */
export function parseCompletionStatus (stdout: string): CompletionStatus | undefined {
  try {
    const value = JSON.parse(stdout) as Partial<CompletionStatus> | null
    if (typeof value?.wanted !== 'boolean') return undefined

    return { shell: typeof value.shell === 'string' ? value.shell : '', wanted: value.wanted, answer: typeof value.answer === 'string' ? value.answer : '' }
  } catch {
    return undefined
  }
}
