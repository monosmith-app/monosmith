import { describe, expect, it } from 'vitest'
import { classifyFolder } from './classify-folder.policy.ts'

describe('classifyFolder', () => {
  it('calls a folder with nothing in it empty', () => {
    expect(classifyFolder([])).toBe('empty')
  })

  it('ignores the repository, the editor settings and operating system junk', () => {
    expect(classifyFolder(['.git', '.vscode', '.DS_Store', 'Thumbs.db', 'desktop.ini'])).toBe('empty')
  })

  it('counts any other file or hidden folder as content', () => {
    expect(classifyFolder(['.git', 'README.md'])).toBe('unrecognised')
    expect(classifyFolder(['.idea'])).toBe('unrecognised')
    expect(classifyFolder(['.gitignore'])).toBe('unrecognised')
    expect(classifyFolder(['src'])).toBe('unrecognised')
  })
})
