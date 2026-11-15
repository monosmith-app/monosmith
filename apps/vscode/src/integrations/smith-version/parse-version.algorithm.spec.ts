import { describe, expect, it } from 'vitest'
import { parseVersion } from './parse-version.algorithm.ts'

describe('parseVersion', () => {
  it('reads the version out of the --version line', () => {
    expect(parseVersion('smith version 1.56.0\n')).toBe('1.56.0')
  })

  it('keeps a pre-release tag', () => {
    expect(parseVersion('smith version 1.57.0-rc.1')).toBe('1.57.0-rc.1')
  })

  it('has none when the text has no version', () => {
    expect(parseVersion('command not found')).toBeUndefined()
  })
})
