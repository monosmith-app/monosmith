import { describe, expect, it } from 'vitest'
import { parseNeededTools } from './add-needed-tools.mapper.ts'

describe('parseNeededTools', () => {
  it('reads the tools and the problem of each from the lines smith add printed', () => {
    const lines = [
      'Wrote 12 file(s):',
      '! golangci-lint is 2.13.0, not the 2.15.0 that CI pins: smith lint of the go projects fails or gives other results than CI.',
      '! Python is not installed: smith build fails.',
      '→ Run smith tools install golangci-lint python to choose how to get them.',
    ]
    expect(parseNeededTools(lines)).toEqual({
      names:    ['golangci-lint', 'python'],
      problems: ['golangci-lint is 2.13.0, not the 2.15.0 that CI pins: smith lint of the go projects fails or gives other results than CI.', 'Python is not installed: smith build fails.'],
    })
  })

  it('reads the line for one tool', () => {
    expect(parseNeededTools(['! Go is not installed: x.', '→ Run smith tools install go to choose how to get it.'])?.names).toEqual(['go'])
  })

  it('says nothing when no tool is lacking', () => {
    expect(parseNeededTools(['Wrote 3 file(s):', 'Done: app web is ready.'])).toBeUndefined()
  })
})
