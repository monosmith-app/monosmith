/**
 * The shape of `packages/plugin-sdk/schema/protocol.schema.json`, the JSON
 * Schema (draft 2020-12) of the plugin protocol, as far as the SDK reads it:
 * the types, the wire codec and the host's capability check all follow it.
 */
export interface SchemaNode {
  '$ref'?:                 string
  'type'?:                 string | string[]
  'properties'?:           Record<string, SchemaNode>
  'required'?:             string[]
  'items'?:                SchemaNode
  'additionalProperties'?: SchemaNode | boolean
  'enum'?:                 string[]
  'anyOf'?:                SchemaNode[]
  'description'?:          string
  'x-enumDescriptions'?:   Record<string, string>
}

/** One row of the methods table: what a method takes, gives and needs. */
export interface MethodRow {
  description: string
  params:      SchemaNode
  result:      SchemaNode
  /** The capability a plugin advertises to be asked the method; `null` when it is always available. */
  capability:  string | null
}

export interface ErrorCodeRow {
  name:        string
  code:        number
  description: string
}

export interface RoleRow {
  name:         string
  description?: string
  /** `exclusive` (one per project), `additive` (any number) or `perSlot` (one per task name provided). */
  cardinality:  'exclusive' | 'additive' | 'perSlot'
  /** The kinds of plugin that play the role when the plugin names none. */
  defaultFor:   string[]
}

export interface SchemaDocument {
  '$schema':           string
  '$id':               string
  'title':             string
  'description':       string
  'x-protocolVersion': number
  'x-transport': {
    framing:         string
    encoding:        string
    jsonrpc:         string
    maxMessageBytes: number
  }
  'x-methods':    Record<string, MethodRow>
  'x-errorCodes': ErrorCodeRow[]
  'x-roles':      RoleRow[]
  '$defs':        Record<string, SchemaNode>
}
