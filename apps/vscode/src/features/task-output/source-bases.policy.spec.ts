import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import { sourceBases } from './source-bases.policy.ts'

const folders = [
  { root: join('/r', 'a'), projects: [{ name: 'web', dir: 'apps/web' }] },
  { root: join('/r', 'b'), projects: [{ name: 'web', dir: 'web' }, { name: 'api', dir: 'api' }] },
]

describe('sourceBases', () => {
  it('tries the project directory first, then the folder', () => {
    expect(sourceBases(folders, 'api:build', join('/r', 'b'))).toEqual([join('/r', 'b', 'api'), join('/r', 'b')])
  })

  it('tries every folder that has the project when the folder is not known', () => {
    expect(sourceBases(folders, 'web:lint')).toEqual([join('/r', 'a', 'apps', 'web'), join('/r', 'b', 'web'), join('/r', 'a'), join('/r', 'b')])
  })

  it('tries only the folders for a line that names no task or an unknown project', () => {
    expect(sourceBases(folders, undefined)).toEqual([join('/r', 'a'), join('/r', 'b')])
    expect(sourceBases(folders, 'nope:build', join('/r', 'a'))).toEqual([join('/r', 'a')])
  })
})
