import type { SchemaDocument, SchemaNode } from '../protocol/index.ts'
import { ParamsError } from './params-error.error.ts'
import { baseNode, isNullable, kindOf } from './schema-lookup.algorithm.ts'

const MAX_DEPTH = 64

/** The JSON test and the words for the error of each scalar kind. */
const SCALARS: Readonly<Record<string, { is: (value: unknown) => boolean, article: string }>> = {
  string:  { is: value => typeof value === 'string', article: 'a string' },
  integer: { is: value => typeof value === 'number' && Number.isSafeInteger(value), article: 'an integer' },
  number:  { is: value => typeof value === 'number', article: 'a number' },
  boolean: { is: value => typeof value === 'boolean', article: 'a boolean' },
}

/**
 * Shapes a received value the way the Go SDK's decoder does, so the types the
 * generated interfaces promise hold:
 *
 * - a required member that is absent or `null` is its zero value (`[]`, `{}`,
 *   `''`, `0`, `false`): the Go host writes a nil slice as `null`;
 * - an optional member that is `null` is removed;
 * - members the schema does not know stay as they are.
 *
 * It throws {@link ParamsError} where a value has the wrong JSON type. The
 * value is edited in place and returned.
 */
export function normalizeBySchema (schema: SchemaDocument, node: SchemaNode, value: unknown, path: string, depth = 0): unknown {
  if (depth > MAX_DEPTH) throw new ParamsError(`${path} is nested too deeply`)
  if (value === null || value === undefined) return isNullable(node) ? null : zeroOf(schema, node, path, depth)
  const kind = kindOf(schema, node)
  const base = baseNode(schema, node)
  const scalar = SCALARS[kind]
  if (scalar !== undefined) {
    if (!scalar.is(value)) throw wrongType(path, scalar.article, value)

    return value
  }
  if (kind === 'array') {
    if (!Array.isArray(value)) throw wrongType(path, 'an array', value)

    return value.map((item, i) => normalizeBySchema(schema, base.items ?? {}, item, `${path}[${i}]`, depth + 1))
  }
  if (kind === 'map') return normalizeMap(schema, base, value, path, depth)

  return kind === 'struct' ? normalizeStruct(schema, base, value, path, depth) : value
}

function normalizeMap (schema: SchemaDocument, base: SchemaNode, value: unknown, path: string, depth: number): unknown {
  if (!isObject(value)) throw wrongType(path, 'an object', value)
  const values = typeof base.additionalProperties === 'object' ? base.additionalProperties : {}
  const keepsNull = kindOf(schema, values) === 'any'
  for (const [name, member] of Object.entries(value)) {
    if (member !== null) value[name] = normalizeBySchema(schema, values, member, `${path}.${name}`, depth + 1)
    else if (!keepsNull) delete value[name]
  }

  return value
}

function normalizeStruct (schema: SchemaDocument, base: SchemaNode, value: unknown, path: string, depth: number): unknown {
  if (!isObject(value)) throw wrongType(path, 'an object', value)
  const required = new Set(base.required)
  const properties = Object.entries(base.properties ?? {})
  for (const [name, property] of properties) {
    const member = value[name]
    const at = `${path}.${name}`
    if (member !== undefined && member !== null) {
      value[name] = normalizeBySchema(schema, property, member, at, depth + 1)
    } else if (required.has(name)) {
      const zero = zeroOf(schema, property, at, depth + 1)
      if (zero !== undefined) value[name] = zero
    } else {
      delete value[name]
    }
  }

  return value
}

/** The value Go's decoder leaves in a member the message did not carry. */
function zeroOf (schema: SchemaDocument, node: SchemaNode, path: string, depth: number): unknown {
  if (isNullable(node)) return null
  const kind = kindOf(schema, node)
  if (kind === 'string') return ''
  if (kind === 'integer' || kind === 'number') return 0
  if (kind === 'boolean') return false
  if (kind === 'array') return []
  if (kind === 'map') return {}

  return kind === 'struct' ? normalizeBySchema(schema, node, {}, path, depth + 1) : undefined
}

function isObject (value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function wrongType (path: string, expected: string, value: unknown): ParamsError {
  const got = Array.isArray(value) ? 'an array' : (typeof value === 'object' ? 'an object' : `a ${typeof value}`)

  return new ParamsError(`${path} must be ${expected}, not ${got}`)
}
