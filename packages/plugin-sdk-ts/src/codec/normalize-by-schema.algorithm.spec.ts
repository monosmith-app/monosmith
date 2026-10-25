import { describe, expect, it } from 'vitest'
import { decodeParams, decodeResult } from './method-codec.algorithm.ts'
import { ParamsError } from './params-error.error.ts'

describe('decodeParams', () => {
  it('turns the null a Go host writes for a nil slice into an empty array, and an absent required member too', () => {
    expect(decodeParams('discover', { manifests: null })).toEqual({ manifests: [] })
    expect(decodeParams('discover', {})).toEqual({ manifests: [] })
    expect(decodeParams('discover', undefined)).toEqual({ manifests: [] })
    expect(decodeParams('contribute', { projects: [{ name: 'a', root: 'r', language: 'l', kind: 'lib', tasks: null, tags: null }] }))
      .toEqual({ projects: [{ name: 'a', root: 'r', language: 'l', kind: 'lib' }] })
  })

  it('gives a required scalar its zero value and removes a null optional member', () => {
    expect(decodeParams('initialize', { host: null, workspaceRoot: null })).toEqual({ protocolVersions: [], host: { name: '', version: '' }, workspaceRoot: '' })
    expect(decodeParams('dependencies/install', { packages: null, dev: null })).toEqual({})
  })

  it('keeps members it does not know', () => {
    expect(decodeParams('discover', { manifests: ['a'], next: { generation: 2 } })).toEqual({ manifests: ['a'], next: { generation: 2 } })
  })

  it('refuses a value of the wrong JSON type, naming the member', () => {
    expect(() => decodeParams('discover', { manifests: 'a/b.json' })).toThrow(new ParamsError('params.manifests must be an array, not a string'))
    expect(() => decodeParams('initialize', { protocolVersions: [1.5] })).toThrow('params.protocolVersions[0] must be an integer')
    expect(() => decodeParams('discover', [])).toThrow('params must be an object, not an array')
    expect(() => decodeParams('version/write', { project: { name: 4 }, version: '1.0.0' })).toThrow('params.project.name must be a string')
  })
})

describe('decodeResult', () => {
  it('shapes a result like params, so a host reads arrays where Go wrote null', () => {
    expect(decodeResult('doctor', { findings: null })).toEqual({ findings: [] })
    expect(decodeResult('shutdown', undefined)).toEqual({})
  })
})
