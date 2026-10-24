import { describe, expect, it } from 'vitest'
import { judgeSkew, type SkewCheck } from './version-skew.policy.ts'

const check = (over: Partial<SkewCheck> = {}): SkewCheck => ({
  editorSmith:     '/repo/node_modules/.bin/smith',
  editorVersion:   '1.54.10',
  editorSource:    'workspace',
  terminalSmith:   '/usr/bin/smith',
  terminalVersion: '1.56.0',
  ...over,
})

describe('judgeSkew', () => {
  it('warns when the workspace smith differs from the PATH smith', () => {
    const warning = judgeSkew(check())
    expect(warning).toContain('1.54.10')
    expect(warning).toContain('1.56.0')
  })

  it('is quiet when the versions are the same', () => {
    expect(judgeSkew(check({ terminalVersion: '1.54.10' }))).toBeUndefined()
  })

  it('is quiet when the editor runs the PATH smith', () => {
    expect(judgeSkew(check({ editorSource: 'path', editorSmith: '/usr/bin/smith' }))).toBeUndefined()
  })

  it('is quiet when there is no smith on the PATH or its version is unknown', () => {
    expect(judgeSkew(check({ terminalSmith: undefined }))).toBeUndefined()
    expect(judgeSkew(check({ terminalVersion: undefined }))).toBeUndefined()
  })

  it('warns about a setting that differs too', () => {
    expect(judgeSkew(check({ editorSource: 'setting' }))).toContain('monosmith.path')
  })
})
