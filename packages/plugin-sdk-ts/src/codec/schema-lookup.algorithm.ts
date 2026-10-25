import type { SchemaDocument, SchemaNode } from '../protocol/index.ts'

/** What a schema node holds, as far as the codec cares. */
export type NodeKind = 'string' | 'integer' | 'number' | 'boolean' | 'array' | 'struct' | 'map' | 'any'

/** Follows `$ref` to the definition it names. */
export function resolveNode (schema: SchemaDocument, node: SchemaNode): SchemaNode {
  let current = node
  for (let hops = 0; current.$ref !== undefined; hops++) {
    const name = current.$ref.slice('#/$defs/'.length)
    const target = schema.$defs[name]
    if (target === undefined || hops > 16) throw new Error(`the schema has no definition ${current.$ref}`)
    current = target
  }

  return current
}

/** Whether `null` is a value of the node: `["integer", "null"]`, or an `anyOf` with null. */
export function isNullable (node: SchemaNode): boolean {
  return Array.isArray(node.type) ? node.type.includes('null') : node.anyOf?.some(branch => branch.type === 'null') ?? false
}

/** The node without its `null` alternative. */
export function withoutNull (schema: SchemaDocument, node: SchemaNode): SchemaNode {
  if (Array.isArray(node.type)) {
    const types = node.type.filter(t => t !== 'null')

    return { ...node, type: types.length === 1 ? types[0] : types }
  }
  const branches = node.anyOf?.filter(branch => branch.type !== 'null')

  return branches !== undefined && branches.length === 1 ? resolveNode(schema, branches[0]) : node
}

/** The kind of value a node describes, after `$ref` and null are set aside. */
export function kindOf (schema: SchemaDocument, node: SchemaNode): NodeKind {
  const base = resolveNode(schema, withoutNull(schema, node))
  switch (base.type) {
    case 'string': { return 'string'
    }
    case 'integer': { return 'integer'
    }
    case 'number': { return 'number'
    }
    case 'boolean': { return 'boolean'
    }
    case 'array': { return 'array'
    }
    case 'object': { return base.properties === undefined ? 'map' : 'struct'
    }
    default: { return 'any'
    }
  }
}

/** The node as the codec walks it: nullability removed, references followed. */
export function baseNode (schema: SchemaDocument, node: SchemaNode): SchemaNode {
  return resolveNode(schema, withoutNull(schema, node))
}
