import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import { folderOfFile, toFolderPath } from './folder-path.algorithm.ts'

const root = join(process.cwd(), 'ws')

describe('toFolderPath', () => {
  it('gives the path inside the folder with forward slashes', () => {
    expect(toFolderPath(root, join(root, 'apps', 'web', 'tsconfig.json'))).toBe('apps/web/tsconfig.json')
  })

  it('gives nothing for the folder itself or a file outside it', () => {
    expect(toFolderPath(root, root)).toBeUndefined()
    expect(toFolderPath(root, join(root, '..', 'other', 'a.json'))).toBeUndefined()
  })
})

describe('folderOfFile', () => {
  it('picks the deepest folder that holds the file', () => {
    const inner = join(root, 'inner')
    expect(folderOfFile([root, inner], join(inner, 'a.json'))).toBe(inner)
    expect(folderOfFile([root, inner], join(root, 'a.json'))).toBe(root)
  })

  it('finds no folder for a file elsewhere', () => {
    expect(folderOfFile([root], join(root, '..', 'x.json'))).toBeUndefined()
  })
})
