import { describe, expect, it } from 'vitest'
import { findSmithRoots } from './find-smith-roots.policy.ts'

describe('findSmithRoots', () => {
  it('keeps the folders that have a smith.toml', () => {
    const folders = [{ root: '/a', name: 'a' }, { root: '/b', name: 'b' }, { root: '/c', name: 'c' }]
    const roots = findSmithRoots(folders, (path) => path === '/a/smith.toml' || path === '/c/smith.toml')
    expect(roots.map((folder) => folder.name)).toEqual(['a', 'c'])
  })

  it('reads Windows paths', () => {
    const roots = findSmithRoots([{ root: String.raw`C:\repo`, name: 'repo' }], (path) => path === String.raw`C:\repo\smith.toml`)
    expect(roots).toHaveLength(1)
  })

  it('finds none when no folder has one', () => {
    expect(findSmithRoots([{ root: '/a', name: 'a' }], () => false)).toEqual([])
  })
})
