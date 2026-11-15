import { posix, win32 } from 'node:path'
import { describe, expect, it } from 'vitest'
import { locateSmith, type SmithSearch } from './locate-smith.policy.ts'

const root = '/work/repo'
const own = posix.join(root, 'node_modules', '.bin', 'smith')

function search (existing: string[], over: Partial<SmithSearch> = {}): SmithSearch {
  return {
    setting:       '',
    workspaceRoot: root,
    trusted:       true,
    platform:      'linux',
    pathVariable:  '/usr/local/bin:/usr/bin',
    exists:        (path) => existing.includes(path),
    ...over,
  }
}

describe('locateSmith', () => {
  it('prefers the setting, even when it names nothing that exists', () => {
    expect(locateSmith(search([], { setting: ' /opt/smith ' })).location).toEqual({ executable: '/opt/smith', source: 'setting' })
  })

  it('prefers the workspace install to the PATH', () => {
    expect(locateSmith(search([own, '/usr/bin/smith'])).location).toEqual({ executable: own, source: 'workspace' })
  })

  it('finds the pinned binary of the repository', () => {
    const pinned = posix.join(root, '.smith', 'bin', 'smith')
    expect(locateSmith(search([pinned])).location?.executable).toBe(pinned)
  })

  it('falls back to the PATH', () => {
    expect(locateSmith(search(['/usr/bin/smith'])).location).toEqual({ executable: '/usr/bin/smith', source: 'path' })
  })

  it('uses the Windows names and separators on Windows', () => {
    const cmd = win32.join(String.raw`C:\repo`, 'node_modules', '.bin', 'smith.cmd')
    expect(locateSmith(search([cmd], { platform: 'win32', workspaceRoot: String.raw`C:\repo` })).location?.executable).toBe(cmd)
    const onPath = win32.join(String.raw`C:\tools`, 'smith.exe')
    const found = locateSmith(search([onPath], { platform: 'win32', workspaceRoot: String.raw`C:\repo`, pathVariable: String.raw`C:\bin;C:\tools` }))
    expect(found.location?.executable).toBe(onPath)
  })

  it('says there is none', () => {
    const found = locateSmith(search([]))
    expect(found.location).toBeUndefined()
    expect(found.skippedUntrusted).toBe(false)
  })

  it('leaves the folder\'s own smith alone when the folder is not trusted', () => {
    const found = locateSmith(search([own, '/usr/bin/smith'], { trusted: false }))
    expect(found.location).toEqual({ executable: '/usr/bin/smith', source: 'path' })
    expect(found.skippedUntrusted).toBe(true)
  })

  it('finds nothing in an untrusted folder that only has its own smith', () => {
    const found = locateSmith(search([own], { trusted: false }))
    expect(found.location).toBeUndefined()
    expect(found.skippedUntrusted).toBe(true)
  })

  it('still honours the setting in an untrusted folder', () => {
    expect(locateSmith(search([], { trusted: false, setting: '/opt/smith' })).location?.source).toBe('setting')
  })

  it('reports the smith the PATH has, even when another one wins', () => {
    expect(locateSmith(search([own, '/usr/bin/smith'])).pathSmith).toBe('/usr/bin/smith')
  })
})
