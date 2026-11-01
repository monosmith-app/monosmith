import { describe, expect, it } from 'vitest'
import { readDeriveReport } from './derive-report.mapper.ts'

describe('readDeriveReport', () => {
  it('says the feature is off when smith.toml does not enable it', () => {
    expect(readDeriveReport({ code: 1, stdout: '', stderr: '✖ derived files are off in this workspace\n  → set enabled = true under [derive] in smith.toml\n' })).toEqual({ state: 'off', files: [], total: 0 })
  })

  it('shows nothing for a smith that has no derive command', () => {
    expect(readDeriveReport({ code: 1, stdout: '', stderr: 'Error: unknown command "derive" for "smith"\n' }).state).toBe('off')
  })

  it('is current when smith exits 0', () => {
    expect(readDeriveReport({ code: 0, stdout: '12 derived files are up to date.\n', stderr: '' })).toEqual({ state: 'current', files: [], total: 12 })
  })

  it('reads a line for each file that needs action', () => {
    const stdout = [
      '  Drift      packages/web/tsconfig.lib.json  (tsconfig-references)  edited by hand',
      '  Stale      packages/api/jest.config.ts  (jest-display-name)  the project was renamed',
      '5 derived files, 2 need action, 3 up to date or owned by you.',
      '',
    ].join('\n')
    const report = readDeriveReport({ code: 1, stdout, stderr: '✖ 2 derived files are out of date\n' })
    expect(report.state).toBe('pending')
    expect(report.total).toBe(5)
    expect(report.files).toEqual([
      { path: 'packages/web/tsconfig.lib.json', status: 'drift', renderer: 'tsconfig-references', detail: 'edited by hand' },
      { path: 'packages/api/jest.config.ts', status: 'stale', renderer: 'jest-display-name', detail: 'the project was renamed' },
    ])
  })

  it('ignores a line of a status it does not know', () => {
    const report = readDeriveReport({ code: 1, stdout: '  Unknown    a.json  (r)  x\n', stderr: '✖ 1 derived files are out of date\n  → run smith derive\n' })
    expect(report.state).toBe('failed')
    expect(report.message).toBe('1 derived files are out of date Try: run smith derive')
  })

  it('returns smith words when it fails', () => {
    expect(readDeriveReport({ code: 1, stdout: '', stderr: '✖ smith.toml is not valid\n  → fix line 3\n' })).toMatchObject({ state: 'failed', message: 'smith.toml is not valid Try: fix line 3' })
  })

  it('names the exit code when smith says nothing', () => {
    expect(readDeriveReport({ code: 2, stdout: '', stderr: '' }).message).toBe('smith stopped with exit code 2')
  })
})
