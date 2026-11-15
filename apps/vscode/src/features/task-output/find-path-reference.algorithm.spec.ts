import { describe, expect, it } from 'vitest'
import { findPathReferences } from './find-path-reference.algorithm.ts'

describe('findPathReferences', () => {
  it('finds path:line:col and says where in the line it is', () => {
    const line = 'web:build | src/app/main.ts:12:5: boom'
    const [found] = findPathReferences(line)
    expect(found).toMatchObject({ file: 'src/app/main.ts', line: 12, column: 5 })
    expect(line.slice(found?.start, found?.end)).toBe('src/app/main.ts:12:5')
  })

  it('finds path:line without a column and tsc style path(line,col)', () => {
    expect(findPathReferences('see a/b.go:7 and c.ts(3,4): error')).toMatchObject([
      { file: 'a/b.go', line: 7 },
      { file: 'c.ts', line: 3, column: 4 },
    ])
    expect(findPathReferences('a/b.go:7')[0]?.column).toBeUndefined()
  })

  it('finds a Windows path with a drive letter and a relative one with dots', () => {
    expect(findPathReferences(String.raw`at C:\repo\src\a.ts:9:1`)).toMatchObject([{ file: String.raw`C:\repo\src\a.ts`, line: 9, column: 1 }])
    expect(findPathReferences('./main.go:12:5: x')).toMatchObject([{ file: './main.go', line: 12 }])
    expect(findPathReferences('../lib/x.py:2: y')).toMatchObject([{ file: '../lib/x.py', line: 2 }])
  })

  it('finds several in one line', () => {
    expect(findPathReferences('a.ts:1:1 conflicts with b.ts:2:2')).toHaveLength(2)
  })

  it('does not take a url or a bare file name for a place', () => {
    expect(findPathReferences('listening on http://localhost:3000/x.js:4')).toEqual([])
    expect(findPathReferences('wrote dist/main.js')).toEqual([])
    expect(findPathReferences('took 12:30:45')).toEqual([])
  })
})
