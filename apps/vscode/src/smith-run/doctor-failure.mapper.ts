import { say } from '../words/index.ts'
/** A project that smith wrote and then found failing its own `smith doctor` checks. */
export interface DoctorFailure {
  /** The project. */
  name:         string
  /** How many files smith wrote; 0 when unknown. */
  files:        number
  /** What could not be checked or failed, one short sentence each ("pkg-config does not find webkit2gtk-4.1"). */
  reasons:      string[]
  /** The block smith printed with its steps, whole, to keep in the Output channel. */
  steps:        string[]
  /** Whether the commit that was asked for happened. */
  committed:    boolean
  /** The sentence in which smith says why it did not commit; empty when it said none. */
  notCommitted: string
}

/**
 * The text of the notice: what was made, what could not be checked and why,
 * and whether the commit happened. It says "created", not "refused".
 * @param failure - what smith printed
 * @returns one paragraph
 */
export function describeDoctorFailure (failure: DoctorFailure): string {
  const files = failure.files > 0 ? ` (${failure.files} files)` : ''
  const why = failure.reasons.length === 0 ? say('vscode.smith-run.a-check-of-smith-doctor-fails') : failure.reasons.join('; ')
  const commit = say(failure.committed ? 'vscode.smith-run.it-was-committed' : 'vscode.smith-run.nothing-was-committed-the-files-are')

  return say('vscode.smith-run.monosmith-the-project-was-created-but', { name: failure.name, files, why, commit })
}

/**
 * Reads the output of a failed `smith add` and tells whether the project was
 * written and then failed its doctor checks (a tool or a system library is
 * missing). Such a project stays on disk, and smith holds back the commit.
 * @param output - what smith printed, standard output then standard error
 * @returns the facts, or undefined for any other failure
 */
export function parseDoctorFailure (output: string): DoctorFailure | undefined {
  const name = /the new project (\S+) was written but fails smith doctor/.exec(output)?.[1]
  if (name === undefined) return undefined

  const lines = output.split(/\r?\n/)
  const start = lines.findIndex((line) => line.startsWith(say('vscode.smith-run.the-new-project-fails-smith-doctor', { name })))
  const rest = start === -1 ? [] : lines.slice(start)
  const block: string[] = []
  for (const line of rest) {
    if (block.length > 0 && line.trim() !== '' && !/^\s/.test(line)) break
    block.push(line.trimEnd())
  }
  const steps = block.filter((line, index, all) => line !== '' || index < all.length - 1)
  const reasons = block.filter((line) => line.trimStart().startsWith('✖')).map((line) => line.trimStart().slice(1).trim())

  return {
    name,
    files:        Number(/Wrote (\d+) file/.exec(output)?.[1] ?? 0),
    reasons,
    steps,
    committed:    /^Committed\b/m.test(output),
    notCommitted: /^Not committed.*$/m.exec(output)?.[0]?.trim() ?? '',
  }
}
