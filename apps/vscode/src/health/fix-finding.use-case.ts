import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import type { DoctorFinding } from './doctor-report.contract.ts'

/** What fixing one check needs from the editor. */
export interface FixFindingDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:     (args: string[]) => Promise<SmithRun>
  /** Asks a question of the fix; answers the index of the picked option, or undefined when the person cancels. */
  pick:    (question: string, labels: string[]) => Promise<number | undefined>
  /** Shows what the fix would change (smith's review, with the plan of each command) and answers whether to apply it. */
  confirm: (review: string) => Promise<boolean>
}

/** How fixing one check ended: applied (`lines` are what smith printed; `incomplete` when other checks still fail), refused with smith's words, or declined. */
export type FixFindingOutcome =
  | { kind: 'fixed', lines: string[], incomplete: boolean } |
  { kind: 'refused', message: string } |
  { kind: 'declined' } |
  { kind: 'nothing' }

/**
 * Fixes one failed check the way `smith doctor --fix --only` does: asks the
 * fix's questions, shows smith's review (the diffs, the commands and the plan
 * each command prints), and only on a yes applies it with `--yes`. Nothing is
 * decided silently: every answer is the person's.
 * @param deps - how to run smith and ask the person
 * @param finding - the failed check that carries the fix
 * @returns how it ended; it never throws
 */
export async function fixFinding (deps: FixFindingDeps, finding: DoctorFinding): Promise<FixFindingOutcome> {
  const fix = finding.fix
  if (fix === undefined || finding.ok) return { kind: 'nothing' }

  const answers: number[] = []
  const questions = fix.choices ?? []
  for (const choice of questions) {
    const picked = await deps.pick(choice.question, choice.options.map((option) => option.label))
    if (picked === undefined) return { kind: 'declined' }
    answers.push(picked + 1)
  }
  const args = ['doctor', '--fix', '--only', fix.description, ...(answers.length > 0 ? ['--choose', answers.join(',')] : [])]

  const review = await deps.run(args)
  // The review of an unsafe fix ends non-zero when nobody could answer; it is still the review.
  // What follows "not applied" tells a terminal to run a command: here the Apply button is that command.
  const shown = review.stdout.split(/\r?\n(?=not applied:)/, 1)[0]?.trim() ?? ''
  if (shown === '') return { kind: 'refused', message: describeFailure(review) }
  if (!await deps.confirm(shown)) return { kind: 'declined' }

  const done = await deps.run([...args, '--yes'])
  const lines = done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean)
  // doctor ends with a failure status while another check that has no fix still fails.
  const stillFailing = done.code !== 0 && lines.some((line) => line.startsWith('after fixing'))

  return stillFailing || done.code === 0 ? { kind: 'fixed', lines, incomplete: stillFailing } : { kind: 'refused', message: describeFailure(done) }
}
