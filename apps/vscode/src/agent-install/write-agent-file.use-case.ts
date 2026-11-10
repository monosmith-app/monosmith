import { describeFailure, type SmithRun } from '../smith-run/index.ts'

/** What writing the agent file needs from the editor. */
export interface AgentFileDeps {
  /** Runs smith in the workspace folder with these arguments. */
  run:     (args: string[]) => Promise<SmithRun>
  /** Says what is missing or out of date and answers whether to write. */
  confirm: (outOfDate: string[]) => Promise<boolean>
}

/** How writing `AGENTS.md` ended: written (`lines` are smith's `wrote ...` lines), already `current`, `declined`, or refused with smith's words. */
export type AgentFileOutcome =
  | { kind: 'written', lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'current' | 'declined' }

function linesOf (text: string): string[] {
  return text.split(/\r?\n/).map((line) => line.trim()).filter(Boolean)
}

/**
 * Writes `AGENTS.md` (and the short `CLAUDE.md` region that imports it) with
 * `smith agent file`. It first runs `--check`, which writes nothing: a file
 * that is current needs no question, and one that is not is named to the
 * person, who agrees before smith writes. Text outside smith's markers stays.
 * @param deps - how to run smith and ask the person
 * @param confirmed - the person already agreed (a command argument), so the question is skipped
 * @returns what happened; it never throws
 */
export async function writeAgentFile (deps: AgentFileDeps, confirmed = false): Promise<AgentFileOutcome> {
  const check = await deps.run(['agent', 'file', '--check'])
  if (check.code === 0) return { kind: 'current' }

  const outOfDate = linesOf(check.stdout)
  // `--check` fails for a missing or old file and prints what is wrong on stdout; with nothing there, smith could not check at all.
  if (outOfDate.length === 0) return { kind: 'refused', message: describeFailure(check) }
  if (!confirmed && !await deps.confirm(outOfDate)) return { kind: 'declined' }

  const write = await deps.run(['agent', 'file'])

  return write.code === 0 ? { kind: 'written', lines: linesOf(write.stdout) } : { kind: 'refused', message: describeFailure(write) }
}
