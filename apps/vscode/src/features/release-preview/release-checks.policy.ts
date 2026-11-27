import { sayCount, say } from '../../shared/words/index.ts'

/** What the checks before a release are judged from: what the editor already knows. */
export interface ReleaseFacts {
  /** The doctor's failed checks and warnings; undefined while doctor has not run. */
  findings?: number
  /** How the newest test run of the session ended. */
  tests:     'passed' | 'failed' | 'none'
}

/** One check before a release: what it says, and whether it warns. */
export interface ReleaseCheck {
  text: string
  warn: boolean
}

/**
 * The checks the terminal interface shows before a release (`releaseconfirm.Preflight`), as far as
 * the editor knows them: Health and the tests of the session. The branch check needs git facts
 * that smith serve does not give, so it is left out here.
 * @param facts - what the editor knows
 * @returns the checks, Health first
 */
export function releaseChecks (facts: ReleaseFacts): ReleaseCheck[] {
  const doctor: ReleaseCheck = facts.findings === undefined
    ? { text: say('tui.release_confirm.doctor_unknown'), warn: true }
    : (facts.findings > 0 ? { text: sayCount('tui.release_confirm.doctor_findings', facts.findings), warn: true } : { text: say('tui.release_confirm.doctor_clean'), warn: false })
  const tests: Record<ReleaseFacts['tests'], ReleaseCheck> = {
    passed: { text: say('tui.release_confirm.tests_passed'), warn: false },
    failed: { text: say('tui.release_confirm.tests_failed'), warn: true },
    none:   { text: say('tui.release_confirm.tests_none'), warn: true },
  }

  return [doctor, tests[facts.tests]]
}
