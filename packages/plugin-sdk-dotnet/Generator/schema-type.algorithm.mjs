// What a node of the schema is, for the generator: a scalar, an enumeration, a message, a list, a map, ...

import { typeName } from './csharp-names.algorithm.mjs'

/**
 * Describes a schema node:
 * `{ kind, nullable?, name?, items?, values? }` where kind is `string`, `integer`, `number`, `boolean`,
 * `enum`, `struct`, `array`, `map`, `object` (a free-form object) or `any`.
 */
export function describe(schema, node) {
  if (node.$ref !== undefined) {
    const definition = node.$ref.slice('#/$defs/'.length)
    const target = schema.$defs[definition]
    return { kind: target.enum !== undefined ? 'enum' : 'struct', name: typeName(definition), definition }
  }
  const type = node.type
  if (Array.isArray(type)) {
    const concrete = type.filter((t) => t !== 'null')
    return { ...describe(schema, { type: concrete[0] }), nullable: type.includes('null') }
  }
  switch (type) {
    case 'string':
    case 'integer':
    case 'number':
    case 'boolean':
      return { kind: type }
    case 'array':
      return { kind: 'array', items: describe(schema, node.items ?? {}) }
    case 'object':
      return typeof node.additionalProperties === 'object'
        ? { kind: 'map', values: describe(schema, node.additionalProperties) }
        : { kind: 'object' }
    default:
      return { kind: 'any' }
  }
}

/** The C# type of a described node; `optional` marks a message member that may be absent. */
export function csharpType(desc, optional = false) {
  switch (desc.kind) {
    case 'string':
      return 'string'
    case 'integer':
      return desc.nullable ? 'long?' : 'long'
    case 'number':
      return 'double'
    case 'boolean':
      return 'bool'
    case 'enum':
      return desc.name
    case 'struct':
      return optional ? `${desc.name}?` : desc.name
    case 'array':
      return `IReadOnlyList<${csharpType(desc.items)}>`
    case 'map':
      return `IReadOnlyDictionary<string, ${csharpType(desc.values)}>`
    case 'object':
      return 'JsonMap'
    default:
      return 'JsonAny?'
  }
}

/** What an absent member of a type is: the zero value, as Go's decoder leaves it. */
export function csharpDefault(desc) {
  switch (desc.kind) {
    case 'string':
      return 'string.Empty'
    case 'array':
      return '[]'
    case 'map':
      return `new Dictionary<string, ${csharpType(desc.values)}>()`
    case 'object':
      return 'new()'
    case 'struct':
      return 'new()'
    default:
      return undefined
  }
}
