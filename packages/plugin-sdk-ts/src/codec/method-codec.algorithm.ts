import { protocolSchema } from '../protocol/index.ts'
import type { MethodRow } from '../protocol/index.ts'
import { encodeBySchema } from './encode-by-schema.algorithm.ts'
import { encodeJsonValue } from './encode-json-value.algorithm.ts'
import { normalizeBySchema } from './normalize-by-schema.algorithm.ts'

type Side = 'params' | 'result'

function rowOf (method: string): MethodRow | undefined {
  return Object.hasOwn(protocolSchema['x-methods'], method) ? protocolSchema['x-methods'][method] : undefined
}

/** The params of a request, written as the Go SDK writes them; a method the protocol does not know is written as plain JSON. */
export function encodeParams (method: string, params: unknown): string {
  return encode(method, 'params', params)
}

/** The result of a method, written as the Go SDK writes it. */
export function encodeResult (method: string, result: unknown): string {
  return encode(method, 'result', result)
}

/** The params of a received request, shaped as the generated types promise (see normalizeBySchema). Absent params are an empty object. */
export function decodeParams (method: string, params: unknown): unknown {
  return decode(method, 'params', params ?? {})
}

/** The result of a received response, shaped as the generated types promise. */
export function decodeResult (method: string, result: unknown): unknown {
  return decode(method, 'result', result ?? {})
}

function encode (method: string, side: Side, value: unknown): string {
  const schema = rowOf(method)?.[side]

  return schema === undefined || schema === null ? encodeJsonValue(value, side) : encodeBySchema(protocolSchema, schema, value, side)
}

function decode (method: string, side: Side, value: unknown): unknown {
  const schema = rowOf(method)?.[side]

  return schema === undefined || schema === null ? value : normalizeBySchema(protocolSchema, schema, value, side)
}
