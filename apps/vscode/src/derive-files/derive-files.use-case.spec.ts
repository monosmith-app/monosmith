import { describe, expect, it } from 'vitest'
import { changeDerivedFile, explainDerivedFile, readDeriveStatus } from './derive-files.use-case.ts'

describe('readDeriveStatus', () => {
  it('runs the check and reads it', async () => {
    const calls: string[][] = []
    const report = await readDeriveStatus(async (args) => {
      calls.push(args)

      return { code: 0, stdout: '3 derived files are up to date.\n', stderr: '' }
    })
    expect(calls).toEqual([['derive', '--check']])
    expect(report).toEqual({ state: 'current', files: [], total: 3 })
  })
})

describe('explainDerivedFile', () => {
  it('returns what smith printed', async () => {
    expect(await explainDerivedFile(async () => ({ code: 0, stdout: 'a.json\n  owner smith\n', stderr: '' }), 'a.json')).toBe('a.json\n  owner smith\n')
  })

  it('throws smith words with the fix when smith does not derive the file', async () => {
    await expect(explainDerivedFile(async () => ({ code: 1, stdout: '', stderr: '✖ smith does not derive x.json\n  → smith derive --check lists the files it derives\n' }), 'x.json'))
      .rejects.toThrow('smith does not derive x.json Try: smith derive --check lists the files it derives')
  })
})

describe('changeDerivedFile', () => {
  it('passes the action and the file', async () => {
    const calls: string[][] = []
    const change = await changeDerivedFile(async (args) => {
      calls.push(args)

      return { code: 0, stdout: '  ejected   a.json\n', stderr: '' }
    }, 'eject', 'a.json')
    expect(change).toEqual({ ok: true, output: '  ejected   a.json\n' })
    expect(calls).toEqual([['derive', '--eject=a.json', '--yes']])
  })

  it('returns the words of smith when it refuses', async () => {
    expect(await changeDerivedFile(async () => ({ code: 1, stdout: '', stderr: '✖ a.json cannot be adopted\n  → use --replace or --eject\n' }), 'adopt', 'a.json'))
      .toEqual({ ok: false, message: 'a.json cannot be adopted Try: use --replace or --eject' })
  })
})
