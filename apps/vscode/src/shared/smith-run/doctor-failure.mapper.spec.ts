import { describe, expect, it } from 'vitest'
import { describeDoctorFailure, parseDoctorFailure } from './doctor-failure.mapper.ts'

const STDOUT = [
  'Creating app web in apps/web/: a Tauri app',
  'Wrote 32 file(s)',
  'The new project web fails smith doctor:',
  '  × pkg-config does not find webkit2gtk-4.1',
  '      → run: sudo apt-get install libwebkit2gtk-4.1-dev',
  '        then open a new terminal',
  'Not committed: add app web was written, but its check failed, and nobody could choose. Run smith commit --only-operation.',
  '',
].join('\n')
const STDERR = 'smith: the new project web was written but fails smith doctor; follow the steps above, then run smith doctor\n'

describe('parseDoctorFailure', () => {
  const found = parseDoctorFailure(`${STDOUT}\n${STDERR}`)

  it('reads the project, the files and what failed', () => {
    expect(found).toMatchObject({ name: 'web', files: 32, reasons: ['pkg-config does not find webkit2gtk-4.1'] })
  })

  it('keeps the whole block of steps for the Output channel', () => {
    expect(found?.steps).toEqual([
      'The new project web fails smith doctor:',
      '  × pkg-config does not find webkit2gtk-4.1',
      '      → run: sudo apt-get install libwebkit2gtk-4.1-dev',
      '        then open a new terminal',
    ])
  })

  it('says the commit did not happen, with the words of smith', () => {
    expect(found?.committed).toBe(false)
    expect(found?.notCommitted).toContain('Not committed')
  })

  it('knows a commit that happened', () => {
    expect(parseDoctorFailure(`${STDOUT}\nCommitted abc123: add app web\n${STDERR}`)?.committed).toBe(true)
  })

  it('words the notice with the facts: created, how many files, what failed, and that nothing was committed', () => {
    const text = found === undefined ? '' : describeDoctorFailure(found)
    expect(text).toContain('the project web was created (32 files)')
    expect(text).toContain('pkg-config does not find webkit2gtk-4.1')
    expect(text).toContain('Nothing was committed')
    expect(text).not.toMatch(/refused|cancelled/)
  })

  it('leaves any other failure to the caller', () => {
    expect(parseDoctorFailure('× the build failed\n')).toBeUndefined()
    expect(parseDoctorFailure('dotnet is not installed or not on PATH')).toBeUndefined()
  })
})
