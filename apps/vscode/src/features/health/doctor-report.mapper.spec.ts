import { describe, expect, it } from 'vitest'
import { sayDoctor, summarizeDoctor } from './doctor-report.mapper.ts'

describe('summarizeDoctor', () => {
  it('counts passed, warnings and failures apart', () => {
    const summary = summarizeDoctor({
      findings: [
        { check: 'a', ok: true },
        { check: 'b', ok: false, warning: true, detail: 'tracked files' },
        { check: 'c', ok: false, detail: 'CI is out of date', remedy: 'run smith ci generate' },
      ],
    })
    expect(summary).toMatchObject({ passed: 1, warnings: 1, failed: 1 })
    expect(summary.lines[0]).toBe('failed: c')
    expect(summary.lines).toContain('  fix: run smith ci generate')
    expect(summary.lines.indexOf('warning: b')).toBeGreaterThan(summary.lines.indexOf('failed: c'))
  })

  it('has no lines when everything passed', () => {
    expect(summarizeDoctor({ findings: [{ check: 'a', ok: true }] }).lines).toEqual([])
  })

  it('copes with a report without findings', () => {
    expect(summarizeDoctor({})).toMatchObject({ passed: 0, failed: 0 })
  })
})

describe('sayDoctor', () => {
  it('says all passed', () => {
    expect(sayDoctor({ passed: 4, warnings: 0, failed: 0, lines: [] })).toBe('Doctor: all 4 checks passed.')
  })

  it('counts what did not', () => {
    expect(sayDoctor({ passed: 4, warnings: 1, failed: 2, lines: [] })).toBe('Doctor: 2 failed, 1 warnings, 4 passed.')
  })
})
