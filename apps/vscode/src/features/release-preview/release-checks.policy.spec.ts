import { describe, expect, it } from 'vitest'
import { releaseChecks } from './release-checks.policy.ts'

describe('releaseChecks', () => {
  it('passes with a clean doctor and passing tests', () => {
    expect(releaseChecks({ findings: 0, tests: 'passed' })).toEqual([
      { text: 'Health: no findings', warn: false },
      { text: 'Tests: the last test run passed', warn: false },
    ])
  })

  it('warns about open findings, failed tests and what is not known yet', () => {
    expect(releaseChecks({ findings: 2, tests: 'failed' })).toEqual([
      { text: 'Health: 2 findings are open', warn: true },
      { text: 'Tests: the last test run failed', warn: true },
    ])
    expect(releaseChecks({ tests: 'none' })).toEqual([
      { text: 'Health: not checked yet', warn: true },
      { text: 'Tests: no test run in this session', warn: true },
    ])
    expect(releaseChecks({ findings: 1, tests: 'passed' })[0]?.text).toBe('Health: 1 finding is open')
  })
})
