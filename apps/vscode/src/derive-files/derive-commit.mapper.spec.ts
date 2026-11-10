import { describe, expect, it } from 'vitest'
import { deriveArguments } from './derive-arguments.mapper.ts'
import { changeDerivedFile, previewDerive } from './derive-files.use-case.ts'

describe('deriveArguments with a commit', () => {
  it('adds --commit to every change, and to nothing else', () => {
    expect(deriveArguments('apply', '', true)).toEqual(['derive', '--yes', '--commit'])
    expect(deriveArguments('replace', 'a.json', true)).toEqual(['derive', '--replace=a.json', '--yes', '--commit'])
    expect(deriveArguments('check', '', true)).toEqual(['derive', '--check'])
    expect(deriveArguments('explain', 'a.json', true)).toEqual(['derive', '--explain=a.json'])
  })

  it('previews with --dry-run', () => {
    expect(deriveArguments('preview')).toEqual(['derive', '--dry-run'])
  })
})

describe('previewDerive and changeDerivedFile', () => {
  it('shows the plan and writes nothing', async () => {
    const calls: string[][] = []
    const text = await previewDerive(async (args) => {
      calls.push(args)

      return { code: 0, stdout: 'would write a.json\n', stderr: '' }
    })
    expect(text).toBe('would write a.json\n')
    expect(calls).toEqual([['derive', '--dry-run']])
  })

  it('passes the commit choice to smith', async () => {
    const calls: string[][] = []
    await changeDerivedFile(async (args) => {
      calls.push(args)

      return { code: 0, stdout: '', stderr: '' }
    }, 'apply', undefined, true)
    expect(calls).toEqual([['derive', '--yes', '--commit']])
  })
})
