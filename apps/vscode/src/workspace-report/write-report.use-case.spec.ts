import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import { reportFile, writeReport } from './write-report.use-case.ts'

describe('reportFile', () => {
  it('puts the report in the temp folder with a safe name', () => {
    expect(reportFile('/tmp', 'my app/web')).toBe(join('/tmp', 'monosmith-report-my-app-web.html'))
  })

  it('names a folder with nothing usable in its name', () => {
    expect(reportFile('/tmp', '')).toBe(join('/tmp', 'monosmith-report-workspace.html'))
  })
})

describe('writeReport', () => {
  it('asks smith to write the page to the file', async () => {
    const calls: string[][] = []
    const outcome = await writeReport(async (args) => {
      calls.push(args)

      return { code: 0, stdout: 'wrote r.html\n', stderr: '' }
    }, 'r.html')
    expect(outcome).toEqual({ ok: true, file: 'r.html' })
    expect(calls).toEqual([['report', '-o', 'r.html']])
  })

  it('returns the words of smith when it fails', async () => {
    expect(await writeReport(async () => ({ code: 1, stdout: '', stderr: '✖ cannot write the file\n  → pick a folder you can write to\n' }), 'r.html'))
      .toEqual({ ok: false, message: 'cannot write the file Try: pick a folder you can write to' })
  })
})
