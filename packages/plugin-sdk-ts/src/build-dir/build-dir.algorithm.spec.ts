import { describe, expect, it } from 'vitest'
import { buildDirFromProject, inBuildDir, projectBuildDir } from './build-dir.algorithm.ts'

describe('the build directory', () => {
  it('mirrors the workspace under dist/', () => {
    expect(projectBuildDir('apps/web')).toBe('dist/apps/web')
    expect(projectBuildDir('/libs/core/')).toBe('dist/libs/core')
  })

  it('reaches it from inside the project', () => {
    expect(buildDirFromProject('apps/web')).toBe('../../dist/apps/web')
    expect(buildDirFromProject('tools/a/b')).toBe('../../../dist/tools/a/b')
    expect(buildDirFromProject('web')).toBe('../dist/web')
  })

  it('knows a path inside it', () => {
    expect(inBuildDir('dist')).toBe(true)
    expect(inBuildDir('dist/apps/web')).toBe(true)
    expect(inBuildDir('distant/apps')).toBe(false)
    expect(inBuildDir('apps/dist')).toBe(false)
  })
})
