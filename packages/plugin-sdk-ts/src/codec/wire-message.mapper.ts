import { encodeJsonString } from './encode-json-value.algorithm.ts'
import { MessageParseError } from './message-parse.error.ts'

/** A request or notification as it was received. `id` is `null` for a notification. */
export interface ReceivedRequest {
  id:     number | null
  method: string
  params: unknown
}

/** A response as it was received. */
export interface ReceivedResponse {
  id:     number | null
  result: unknown
  error:  { code: number, message: string } | null
}

/** A request as the Go SDK writes it: `jsonrpc`, `id`, `method`, `params`, the params left out when there are none. */
export function writeRequest (id: number, method: string, paramsJson: string | undefined): string {
  const params = paramsJson === undefined ? '' : `,"params":${paramsJson}`

  return `{"jsonrpc":"2.0","id":${id},"method":${encodeJsonString(method)}${params}}`
}

/** A notification as the Go SDK writes it: a request with no `id`, which gets no response. */
export function writeNotification (method: string, paramsJson: string): string {
  return `{"jsonrpc":"2.0","method":${encodeJsonString(method)},"params":${paramsJson}}`
}

/** A successful response:`id` is `null` only when the request could not be read. */
export function writeResult (id: number | null, resultJson: string): string {
  return `{"jsonrpc":"2.0","id":${id === null ? 'null' : id},"result":${resultJson}}`
}

/** An error response. */
export function writeFailure (id: number | null, code: number, message: string): string {
  return `{"jsonrpc":"2.0","id":${id === null ? 'null' : id},"error":{"code":${code},"message":${encodeJsonString(message)}}}`
}

function parseObject (body: string): Record<string, unknown> {
  let value: unknown
  try {
    value = JSON.parse(body)
  } catch (error) {
    throw new MessageParseError(error instanceof Error ? error.message : String(error))
  }
  if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new MessageParseError('a message is a JSON object')

  return value as Record<string, unknown>
}

function readId (message: Record<string, unknown>): number | null {
  const id = message.id
  if (id === undefined || id === null) return null
  if (typeof id !== 'number' || !Number.isSafeInteger(id)) throw new MessageParseError('id must be an integer')

  return id
}

/** Reads a request frame. A `null` or missing id makes it a notification. */
export function parseRequest (body: string): ReceivedRequest {
  const message = parseObject(body)
  const id = readId(message)
  const method = message.method ?? ''
  if (typeof method !== 'string') throw new MessageParseError('method must be a string')

  return { id, method, params: message.params }
}

/** Reads a response frame. */
export function parseResponse (body: string): ReceivedResponse {
  const message = parseObject(body)
  const id = readId(message)
  const error = message.error
  if (error === undefined || error === null) return { id, result: message.result, error: null }
  if (typeof error !== 'object' || Array.isArray(error)) throw new MessageParseError('error must be an object')
  const { code, message: text } = error as Record<string, unknown>
  if (typeof code !== 'number' || typeof text !== 'string') throw new MessageParseError('error needs a numeric code and a message')

  return { id, result: undefined, error: { code, message: text } }
}
