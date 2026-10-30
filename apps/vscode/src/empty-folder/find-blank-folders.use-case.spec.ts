import { describe, expect, it } from 'vitest'
import { findBlankFolders } from './find-blank-folders.use-case.ts'

const folders = [
  { root: '/empty', name: 'empty' },
  { root: '/files', name: 'files' },
  { root: '/smith', name: 'smith' },
  { root: '/gone', name: 'gone' },
]
const contents: Record<string, string[]> = { '/empty': ['.git', '.DS_Store'], '/files': ['notes.txt'], '/smith': [] }
const list = (root: string): string[] | undefined => contents[root]

describe('findBlankFolders', () => {
  it('tells an empty folder from one with files', () => {
    expect(findBlankFolders(folders.slice(0, 2), new Set(), list).map((folder) => folder.kind)).toEqual(['empty', 'unrecognised'])
  })

  it('leaves out the folders it was told to skip and the ones it cannot list', () => {
    expect(findBlankFolders(folders, new Set(['/smith']), list).map((folder) => folder.root)).toEqual(['/empty', '/files'])
  })
})
