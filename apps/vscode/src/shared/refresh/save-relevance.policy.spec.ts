import { describe, expect, it } from 'vitest'
import { isRelevantSave } from './save-relevance.policy.ts'

describe('isRelevantSave', () => {
  it('counts a source file in the folder', () => {
    expect(isRelevantSave('/repo', '/repo/apps/api/src/main.ts')).toBe(true)
  })

  it('ignores a file outside the folder, and a folder that only shares a prefix', () => {
    expect(isRelevantSave('/repo', '/elsewhere/main.ts')).toBe(false)
    expect(isRelevantSave('/repo', '/repo-two/main.ts')).toBe(false)
  })

  it('ignores installs, builds and git or smith state', () => {
    expect(isRelevantSave('/repo', '/repo/node_modules/x/index.js')).toBe(false)
    expect(isRelevantSave('/repo', '/repo/apps/api/dist/main.js')).toBe(false)
    expect(isRelevantSave('/repo', '/repo/.git/COMMIT_EDITMSG')).toBe(false)
    expect(isRelevantSave('/repo', '/repo/.smith/cache/x')).toBe(false)
  })

  it('reads Windows paths, in any case', () => {
    expect(isRelevantSave(String.raw`c:\repo`, String.raw`C:\Repo\apps\api\main.ts`)).toBe(true)
    expect(isRelevantSave(String.raw`c:\repo`, String.raw`C:\repo\node_modules\x.js`)).toBe(false)
  })
})
