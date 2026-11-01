import { describe, expect, it } from 'vitest'
import { detectRepository } from './detect-repository.policy.ts'

const probe = (files: Record<string, string>) => ({
  exists: (path: string) => Object.hasOwn(files, path),
  read:   (path: string) => files[path],
})

describe('detectRepository', () => {
  it('does not offer a folder that has a smith.toml, even one that cannot be read', () => {
    expect(detectRepository('/repo', probe({ '/repo/nx.json': '{}', '/repo/smith.toml': 'broken = ' }))).toBeUndefined()
  })

  it('recognises an Nx workspace by nx.json', () => {
    expect(detectRepository('/repo', probe({ '/repo/nx.json': '{}' }))).toEqual({ kind: 'nx', marker: 'nx.json' })
  })

  it('recognises a Turborepo workspace by turbo.json', () => {
    expect(detectRepository('/repo', probe({ '/repo/turbo.json': '{}' }))).toEqual({ kind: 'turborepo', marker: 'turbo.json' })
  })

  it('prefers Nx when both are there', () => {
    expect(detectRepository('/repo', probe({ '/repo/nx.json': '{}', '/repo/turbo.json': '{}' }))?.kind).toBe('nx')
  })

  it('recognises other monorepo tools and Go workspaces', () => {
    expect(detectRepository('/repo', probe({ '/repo/pnpm-workspace.yaml': '' }))).toEqual({ kind: 'repository', marker: 'pnpm-workspace.yaml' })
    expect(detectRepository('/repo', probe({ '/repo/go.work': '' }))?.marker).toBe('go.work')
  })

  it('recognises a package.json that lists workspaces, as an array or an object', () => {
    expect(detectRepository('/repo', probe({ '/repo/package.json': '{"workspaces":["apps/*"]}' }))?.marker).toBe('package.json')
    expect(detectRepository('/repo', probe({ '/repo/package.json': '{"workspaces":{"packages":["apps/*"]}}' }))?.kind).toBe('repository')
  })

  it('ignores a package.json without workspaces, an empty list or broken text', () => {
    expect(detectRepository('/repo', probe({ '/repo/package.json': '{"name":"x"}' }))).toBeUndefined()
    expect(detectRepository('/repo', probe({ '/repo/package.json': '{"workspaces":[]}' }))).toBeUndefined()
    expect(detectRepository('/repo', probe({ '/repo/package.json': '{oops' }))).toBeUndefined()
  })

  it('finds nothing in a folder with no marker', () => {
    expect(detectRepository('/repo', probe({}))).toBeUndefined()
  })

  it('reads Windows paths', () => {
    expect(detectRepository(String.raw`C:\repo`, probe({ [String.raw`C:\repo\nx.json`]: '{}' }))?.kind).toBe('nx')
  })
})
