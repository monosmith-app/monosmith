import type { DoctorReport } from '../health/index.ts'

/** The check of `smith doctor` about the agent files. */
export const AGENT_FILES_CHECK = 'The MonoSmith agent files are installed and current'

/**
 * What the doctor says about the agent files: `missing` when its check
 * warns, `current` when it passes, `unknown` when the check is not in the
 * report (smith is older, or `[agent] enabled = false` opted out).
 */
export type AgentFilesState = 'missing' | 'current' | 'unknown'

/** Reads the state of the agent files from a doctor report. */
export function judgeAgentFiles (report: DoctorReport): AgentFilesState {
  const finding = report.findings?.find((candidate) => candidate.check === AGENT_FILES_CHECK)
  if (finding === undefined) return 'unknown'

  return finding.ok ? 'current' : 'missing'
}

/** What decides whether to offer the agent once. */
export interface OfferFacts {
  state:        AgentFilesState
  /** The person already answered the offer for this workspace. */
  alreadyAsked: boolean
}

/** Offers the agent only when its files are missing and the person has not answered before. */
export function shouldOfferAgent (facts: OfferFacts): boolean {
  return facts.state === 'missing' && !facts.alreadyAsked
}
