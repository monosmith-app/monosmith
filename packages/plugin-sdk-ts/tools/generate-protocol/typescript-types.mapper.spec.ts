import { describe, expect, it } from 'vitest'
import type { SchemaDocument } from '../../src/protocol/schema-document.contract.ts'
import { mapSchemaToTypes, typeOf } from './typescript-types.mapper.ts'

describe('typeOf', () => {
  it('maps the JSON Schema nodes the protocol uses', () => {
    expect(typeOf({ type: 'string' })).toBe('string')
    expect(typeOf({ type: 'integer' })).toBe('number')
    expect(typeOf({ type: 'boolean' })).toBe('boolean')
    expect(typeOf({ type: ['integer', 'null'] })).toBe('number | null')
    expect(typeOf({ anyOf: [{ $ref: '#/$defs/Project' }, { type: 'null' }] })).toBe('Project | null')
    expect(typeOf({ type: 'array', items: { $ref: '#/$defs/Task' } })).toBe('Task[]')
    expect(typeOf({ type: 'array', items: { type: ['string', 'null'] } })).toBe('(string | null)[]')
    expect(typeOf({ type: 'object', additionalProperties: { type: 'array', items: { type: 'string' } } })).toBe('Record<string, string[]>')
    expect(typeOf({ type: 'object' })).toBe('Record<string, unknown>')
    expect(typeOf({})).toBe('unknown')
  })
})

describe('mapSchemaToTypes', () => {
  const schema = {
    'x-protocolVersion': 7,
    'x-transport':       { maxMessageBytes: 1024 },
    'x-errorCodes':      [{ name: 'Boom', code: -1, description: 'It broke.' }],
    'x-methods':         {
      'version/write': { description: 'Records.', params: { $ref: '#/$defs/Thing' }, result: { $ref: '#/$defs/Empty' }, capability: 'version' },
      'initialize':    { description: 'First.', params: { $ref: '#/$defs/Thing' }, result: { $ref: '#/$defs/Empty' }, capability: null },
    },
    '$defs': {
      Empty: { description: 'No members.', type: 'object' },
      Kind:  { type: 'string', enum: ['a', 'it\'s'] },
      Thing: {
        description: 'A thing */ with a terminator.',
        type:        'object',
        properties:  { name: { type: 'string', description: 'Its name.' }, applies_to: { type: 'array', items: { type: 'string' } }, kind: { $ref: '#/$defs/Kind' } },
        required:    ['name'],
      },
    },
  } as unknown as SchemaDocument

  it('writes constants, one type per definition and the methods table', () => {
    const source = mapSchemaToTypes(schema)
    expect(source).toContain('export const PROTOCOL_VERSION = 7\n')
    expect(source).toContain('export const MAX_MESSAGE_BYTES = 1024\n')
    expect(source).toContain('  Boom: -1,\n')
    expect(source).toContain("export type Kind = 'a' | 'it\\'s'\n")
    expect(source).toContain('export type Empty = Record<string, never>\n')
    expect(source).toContain('  /** Its name. */\n  name: string\n  applies_to?: string[]\n  kind?: Kind\n')
    expect(source).toContain('/** A thing *\\/ with a terminator. */\nexport interface Thing {')
    expect(source).toContain("  'version/write': {\n    params: Thing\n    result: Empty\n    capability: 'version'\n  }\n")
    expect(source).toContain('  initialize: {\n    params: Thing\n    result: Empty\n    capability: null\n  }\n')
    expect(source.endsWith('\n')).toBe(true)
  })
})
