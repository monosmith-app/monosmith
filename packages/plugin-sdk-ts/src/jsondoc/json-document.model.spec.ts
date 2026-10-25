import { describe, expect, it } from 'vitest'
import { JsonDocumentError } from './json-document.error.ts'
import { JsonDocument } from './json-document.model.ts'

// These are the cases of the Go SDK's jsondoc tests: the same input must give the same bytes.
describe('JsonDocument', () => {
  it('keeps the order of the keys, the text of leaves, and edits in place', () => {
    const doc = JsonDocument.parse(`{
  "name": "@acme/web",
  "version": "1.0.0",
  "nx": {
    "targets": {
      "build": { "executor": "x" }
    }
  },
  "files": ["dist", "!**/*.map"],
  "scripts": {
    "test": "jest",
    "e2e": "a && b <c>"
  },
  "count": 1.50
}
`)
    doc.delete('nx')
    doc.object('scripts')!.setString('build', 'tsc -p tsconfig.lib.json')
    doc.object('scripts')!.setString('test', 'jest --ci')
    doc.object('smith', true)!.setString('kind', 'package')
    expect(doc.render()).toBe(`{
  "name": "@acme/web",
  "version": "1.0.0",
  "files": [
    "dist",
    "!**/*.map"
  ],
  "scripts": {
    "test": "jest --ci",
    "e2e": "a && b <c>",
    "build": "tsc -p tsconfig.lib.json"
  },
  "count": 1.50,
  "smith": {
    "kind": "package"
  }
}
`)
  })

  it('refuses what is not an object, and writes an empty object as {}', () => {
    expect(() => JsonDocument.parse('[1]')).toThrow(new JsonDocumentError('not a JSON object'))
    expect(() => JsonDocument.parse('{"a": ')).toThrow(JsonDocumentError)
    expect(() => JsonDocument.parse('{"a": nope}')).toThrow('invalid JSON value')
    expect(() => JsonDocument.parse('{"a": 1} x')).toThrow('after the object')
    expect(JsonDocument.parse('{"a": {}}').render()).toBe('{\n  "a": {}\n}\n')
    expect(JsonDocument.parse('{}').render()).toBe('{}\n')
  })

  it('makes an object set with setJson editable, and writes other values as they are', () => {
    const doc = JsonDocument.parse('{"name": "x"}')
    doc.setJson('jest', { preset: 'ts-jest', roots: ['<rootDir>/src'] })
    doc.object('jest')!.setString('testEnvironment', 'node')
    doc.setJson('private', true)
    expect(doc.render()).toBe(`{
  "name": "x",
  "jest": {
    "preset": "ts-jest",
    "roots": [
      "<rootDir>/src"
    ],
    "testEnvironment": "node"
  },
  "private": true
}
`)
  })

  it('reads strings, and says no for what is not one', () => {
    const doc = JsonDocument.parse('{"a": "x", "n": 1, "o": {}}')
    expect(doc.getString('a')).toBe('x')
    for (const key of ['n', 'o', 'missing']) expect(doc.getString(key), key).toBeUndefined()
    expect(doc.getValue('n')).toBe(1)
    expect(doc.getValue('o')).toEqual({})
    expect(doc.keys()).toEqual(['a', 'n', 'o'])
    expect(doc.size).toBe(3)
  })

  it('gives a duplicate key its first position and last value, and puts a new key last', () => {
    const doc = JsonDocument.parse('{"a": 1, "b": 2, "a": 3}')
    doc.setString('c', 'x')
    doc.setString('a', 'y')
    expect(doc.render()).toBe('{\n  "a": "y",\n  "b": 2,\n  "c": "x"\n}\n')
  })

  it('formats nested arrays and escapes as Go does', () => {
    const doc = JsonDocument.parse(String.raw`{"matrix": [[1,2],[3,[]],{"a":[ ],"b":"x\"y"}], "s": "<>&\u2028"}`)
    expect(doc.render()).toBe(String.raw`{
  "matrix": [
    [
      1,
      2
    ],
    [
      3,
      []
    ],
    {
      "a": [],
      "b": "x\"y"
    }
  ],
  "s": "<>&\u2028"
}
`)
    doc.setString('t', '\u{2028}<')
    expect(doc.render()).toContain(String.raw`"t": "\u2028<"`)
  })
})
