import { describe, expect, it } from 'vitest'
import { protocolSchema } from '../protocol/index.ts'
import { EncodeError } from './encode-error.error.ts'
import { encodeParams, encodeResult } from './method-codec.algorithm.ts'

describe('encodeResult', () => {
  it('writes the members in the order of the schema, whatever the order they were given in', () => {
    const json = encodeResult('initialize', {
      capabilities:    { doctor: true, discover: true },
      languages:       ['widget'],
      kind:            'language',
      plugin:          { version: '0.1.0', name: 'widget' },
      protocolVersion: 1,
    })
    expect(json).toBe('{"protocolVersion":1,"plugin":{"name":"widget","version":"0.1.0"},"kind":"language","capabilities":{"discover":true,"doctor":true},"languages":["widget"]}')
  })

  it('leaves out an optional member that is absent or holds the zero value Go omits, and keeps a required false', () => {
    const json = encodeResult('doctor', {
      findings: [
        { check: 'a', ok: true, detail: '', warning: false, remedy: undefined },
        { check: 'b', ok: false, detail: 'broken', warning: true },
      ],
    })
    expect(json).toBe('{"findings":[{"check":"a","ok":true},{"check":"b","ok":false,"detail":"broken","warning":true}]}')
    const step = encodeResult('ci', { setup: [{ name: 'x', command: [], blocking: false }] })
    expect(step).toBe('{"setup":[{"name":"x","command":[],"blocking":false}]}')
  })

  it('sorts the keys of a map and escapes HTML characters the way Go does', () => {
    const json = encodeResult('contribute', {
      projects: {
        zeta: {
          tasks: {
            b: { command: ['x<y>&z'] },
            a: {
              command: ['\u{2028}'],
            },
          },
        },
        alpha: {},
      },
    })
    expect(json).toBe(String.raw`{"projects":{"alpha":{},"zeta":{"tasks":{"a":{"command":["\u2028"]},"b":{"command":["x\u003cy\u003e\u0026z"]}}}}}`)
  })

  it('writes a struct member even when it is empty, an empty map or array only when it is required', () => {
    expect(encodeParams('ci', { projects: [], audit: {} })).toBe('{"projects":[],"audit":{}}')
    expect(encodeParams('ci', { projects: [] })).toBe('{"projects":[]}')
    expect(encodeResult('discover', { projects: [{ name: 'a', root: 'r', language: 'l', kind: 'lib', tasks: {}, exports: [] }] }))
      .toBe('{"projects":[{"name":"a","root":"r","language":"l","kind":"lib"}]}')
  })

  it('keeps members of free-form objects, with their keys sorted, and members the schema does not know', () => {
    const json = encodeParams('generate', {
      kind: 'app', name: 'n', directory: 'apps/n', workspace: { name: 'w' }, settings: { b: 1, a: { d: null, c: [1, 'x'] } }, future: true,
    })
    expect(json).toBe('{"kind":"app","name":"n","directory":"apps/n","workspace":{"name":"w"},"settings":{"a":{"c":[1,"x"],"d":null},"b":1},"future":true}')
  })

  it('writes shutdown as an empty object', () => {
    expect(encodeResult('shutdown', {})).toBe('{}')
  })

  it('writes the params of a method the protocol does not know as plain JSON', () => {
    expect(encodeParams('nope', { b: 1, a: [undefined] })).toBe('{"a":[null],"b":1}')
  })
})

describe('encodeResult errors', () => {
  it('names the required member that is missing', () => {
    expect(() => encodeResult('discover', {})).toThrow(new EncodeError('result.projects is required'))
    expect(() => encodeResult('discover', { projects: [{ name: 'a' }] })).toThrow('result.projects[0].root is required')
  })

  it('names a member of the wrong type', () => {
    expect(() => encodeResult('discover', { projects: 'x' })).toThrow('result.projects must be an array, not a string')
    expect(() => encodeResult('generate', { files: [1] })).toThrow('result.files[0] must be a string, not a number')
    expect(() => encodeResult('publish', { published: 'yes' })).toThrow('result.published must be a boolean')
    expect(() => encodeResult('ci', { permissions: [], toolchains: [{ name: 'a', version: 'b', github: [] }] })).toThrow('must be an object')
  })

  it('refuses a null for a member that is not nullable', () => {
    expect(() => encodeResult('discover', { projects: null })).toThrow('cannot be null')
  })
})

describe('the schema the codec follows', () => {
  it('is the one of protocol version 1', () => {
    expect(protocolSchema['x-protocolVersion']).toBe(1)
  })
})
