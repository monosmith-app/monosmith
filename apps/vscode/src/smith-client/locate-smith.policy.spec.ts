import { posix, win32 } from 'node:path'
import { describe, expect, it } from 'vitest'
import { locateSmith, type SmithSearch } from './locate-smith.policy.ts'

const root = '/work/repo'

function search (existing: string[], over: Partial<SmithSearch> = {}): SmithSearch {
  return {
    setting:       '',
    workspaceRoot: root,
    platform:      'linux',
    pathVariable:  '/usr/local/bin:/usr/bin',
    exists:        (path) => existing.includes(path),
    ...over,
  }
}

describe('locateSmith', () => {
  it('prefers the setting, even when it names nothing that exists', () => {
    expect(locateSmith(search([], { setting: ' /opt/smith ' }))).toBe('/opt/smith')
  })

  it('prefers the workspace install to the PATH', () => {
    const own = posix.join(root, 'node_modules', '.bin', 'smith')
    expect(locateSmith(search([own, '/usr/bin/smith']))).toBe(own)
  })

  it('finds the pinned binary of the repository', () => {
    const pinned = posix.join(root, '.smith', 'bin', 'smith')
    expect(locateSmith(search([pinned]))).toBe(pinned)
  })

  it('falls back to the PATH', () => {
    expect(locateSmith(search(['/usr/bin/smith']))).toBe('/usr/bin/smith')
  })

  it('uses the Windows names and separators on Windows', () => {
    const cmd = win32.join(String.raw`C:\repo`, 'node_modules', '.bin', 'smith.cmd')
    expect(locateSmith(search([cmd], { platform: 'win32', workspaceRoot: String.raw`C:\repo` }))).toBe(cmd)
    const onPath = win32.join(String.raw`C:\tools`, 'smith.exe')
    expect(locateSmith(search([onPath], { platform: 'win32', workspaceRoot: String.raw`C:\repo`, pathVariable: String.raw`C:\bin;C:\tools` }))).toBe(onPath)
  })

  it('says there is none', () => {
    expect(locateSmith(search([]))).toBeUndefined()
  })
})
