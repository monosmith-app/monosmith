import type { SmithRun } from './smith-run.contract.ts'

/**
 * Takes the `✖` out of smith's problem so it reads as a sentence, and keeps the way-outs.
 * @param stderr - what smith wrote to standard error
 * @returns the problem, then each fix as `Try: ...`; empty when smith said nothing
 */
export function sayProblem (stderr: string): string {
  const lines = stderr.split(/\r?\n/).map((line) => line.trim()).filter(Boolean)
  const [first = '', ...rest] = lines
  const problem = first.replace(/^✖\s*/, '')

  return [problem, ...rest.map((line) => line.replace(/^→\s*/, 'Try: '))].join(' ')
}

/**
 * Why a run failed, in smith's own words and with its fix. smith writes a
 * problem to standard error; a run with none is described by standard output,
 * and a run with neither by its exit code.
 * @param run - a run that ended with a code other than 0
 */
export function describeFailure (run: SmithRun): string {
  const message = sayProblem(run.stderr.trim() === '' ? run.stdout : run.stderr)

  return message === '' ? `smith stopped with exit code ${run.code}` : message
}
