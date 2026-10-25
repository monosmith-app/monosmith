import type { SchemaDocument, SchemaNode } from '../protocol/index.ts'
import { EncodeError } from './encode-error.error.ts'
import { compareKeys, encodeJsonString, encodeJsonValue } from './encode-json-value.algorithm.ts'
import { baseNode, isNullable, kindOf } from './schema-lookup.algorithm.ts'

/**
 * Writes a value as the JSON the Go SDK would write for the same message:
 * the members of a message in the order the schema lists them (the order of the
 * Go fields), an optional member left out when it is absent or holds the
 * zero value Go's `omitempty` drops (`""`, `0`, `false`, an empty array or
 * map), the keys of a map sorted, HTML characters escaped. So the bytes a TypeScript
 * plugin writes equal the bytes a Go plugin writes for the same answer.
 *
 * It throws {@link EncodeError} for a required member that is missing or a value of the wrong type.
 */
export function encodeBySchema (schema: SchemaDocument, node: SchemaNode, value: unknown, path: string): string {
  const kind = kindOf(schema, node)
  if (value === null || value === undefined) {
    if (isNullable(node) || (kind === 'any' && value === null)) return 'null'
    throw new EncodeError(`${path} is required${value === null ? ' and cannot be null' : ''}`)
  }
  const base = baseNode(schema, node)
  switch (kind) {
    case 'string': {
      if (typeof value !== 'string') throw wrongType(path, 'a string', value)

      return encodeJsonString(value)
    }
    case 'integer': {
      if (typeof value !== 'number' || !Number.isSafeInteger(value)) throw wrongType(path, 'an integer', value)

      return String(value)
    }
    case 'number': {
      if (typeof value !== 'number' || !Number.isFinite(value)) throw wrongType(path, 'a number', value)

      return String(value)
    }
    case 'boolean': {
      if (typeof value !== 'boolean') throw wrongType(path, 'a boolean', value)

      return value ? 'true' : 'false'
    }
    case 'array': {
      if (!Array.isArray(value)) throw wrongType(path, 'an array', value)
      const items = base.items ?? {}

      return `[${value.map((item, i) => encodeBySchema(schema, items, item, `${path}[${i}]`)).join(',')}]`
    }
    case 'map': {
      if (!isPlainObject(value)) throw wrongType(path, 'an object', value)
      const values = typeof base.additionalProperties === 'object' ? base.additionalProperties : {}
      const members = Object.entries(value)
        .filter(([, member]) => member !== undefined)
        .sort(([a], [b]) => compareKeys(a, b))
        .map(([name, member]) => `${encodeJsonString(name)}:${encodeBySchema(schema, values, member, `${path}.${name}`)}`)

      return `{${members.join(',')}}`
    }
    case 'struct': {
      if (!isPlainObject(value)) throw wrongType(path, 'an object', value)

      return encodeStruct(schema, base, value, path)
    }
    default: {
      return encodeJsonValue(value, path)
    }
  }
}

function encodeStruct (schema: SchemaDocument, base: SchemaNode, value: Record<string, unknown>, path: string): string {
  const properties = base.properties ?? {}
  const required = new Set(base.required)
  const members: string[] = []
  for (const [name, property] of Object.entries(properties)) {
    const member = value[name]
    if (!required.has(name) && isOmitted(schema, property, member)) continue
    members.push(`${encodeJsonString(name)}:${encodeBySchema(schema, property, member, `${path}.${name}`)}`)
  }
  // Members the schema does not know (a protocol that grew past this SDK) are kept, after the known ones.
  const known = new Set(Object.keys(properties))
  const unknown = Object.keys(value).filter(name => !known.has(name) && value[name] !== undefined).sort(compareKeys)
  for (const name of unknown) members.push(`${encodeJsonString(name)}:${encodeJsonValue(value[name], `${path}.${name}`)}`)

  return `{${members.join(',')}}`
}

/** Go's `omitempty`: absent, null, or the zero value of a scalar, array or map. A struct is never empty. */
function isOmitted (schema: SchemaDocument, property: SchemaNode, member: unknown): boolean {
  if (member === undefined || member === null) return true
  const kind = kindOf(schema, property)
  if (kind === 'string') return member === ''
  if (kind === 'integer' || kind === 'number') return member === 0
  if (kind === 'boolean') return member === false

  return kind === 'array' ? Array.isArray(member) && member.length === 0 : kind === 'map' && isPlainObject(member) && Object.values(member).every(v => v === undefined)
}

function isPlainObject (value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function wrongType (path: string, expected: string, value: unknown): EncodeError {
  const got = Array.isArray(value) ? 'an array' : (typeof value === 'object' ? 'an object' : `a ${typeof value}`)

  return new EncodeError(`${path} must be ${expected}, not ${got}`)
}
