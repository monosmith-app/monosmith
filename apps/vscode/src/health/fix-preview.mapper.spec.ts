import { describe, expect, it } from 'vitest'
import { parseFixPreview } from './fix-preview.mapper.ts'

const proposed = [
  'smith doctor --fix exited with code 1 and printed no JSON: ✖ the cache directory is ignored by git',
  '    → add `.smith/` to .gitignore',
  '',
  '34 passed, 1 failed, 0 warning(s)',
  '',
  '● ignore the cache directory',
  '--- a/.gitignore',
  '+++ b/.gitignore',
  '@@ -17,3 +17,6 @@',
  '+.smith/',
  '',
  '● pin the node version',
  '+24',
  '',
  'Pass --yes to apply these fixes.',
].join('\n')

describe('parseFixPreview', () => {
  it('lists the fixes smith proposes and keeps their diffs', () => {
    const preview = parseFixPreview(proposed)
    expect(preview?.titles).toEqual(['ignore the cache directory', 'pin the node version'])
    expect(preview?.text.startsWith('● ignore the cache directory')).toBe(true)
    expect(preview?.text).toContain('+.smith/')
    expect(preview?.text).not.toContain('Pass --yes')
  })

  it('is undefined when there is nothing doctor can fix', () => {
    expect(parseFixPreview('35 passed, 0 failed\n\nNothing doctor can fix itself: follow the remedies above.')).toBeUndefined()
  })

  it('reads a block without the closing line', () => {
    expect(parseFixPreview('● one\n+x')?.titles).toEqual(['one'])
  })
})
