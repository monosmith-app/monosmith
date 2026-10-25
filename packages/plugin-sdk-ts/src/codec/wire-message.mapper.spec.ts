import { describe, expect, it } from 'vitest'
import { MessageParseError } from './message-parse.error.ts'
import { parseRequest, parseResponse, writeFailure, writeRequest, writeResult } from './wire-message.mapper.ts'

describe('writing', () => {
  it('writes the envelopes in the Go SDK member order', () => {
    expect(writeRequest(3, 'discover', '{"manifests":[]}')).toBe('{"jsonrpc":"2.0","id":3,"method":"discover","params":{"manifests":[]}}')
    expect(writeRequest(1, 'shutdown', undefined)).toBe('{"jsonrpc":"2.0","id":1,"method":"shutdown"}')
    expect(writeResult(3, '{}')).toBe('{"jsonrpc":"2.0","id":3,"result":{}}')
    expect(writeFailure(null, -32700, 'bad <json>')).toBe(String.raw`{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"bad \u003cjson\u003e"}}`)
  })
})

describe('parseRequest', () => {
  it('reads a request, and a notification as a request without an id', () => {
    expect(parseRequest('{"jsonrpc":"2.0","id":7,"method":"doctor","params":{"projects":[]}}')).toEqual({ id: 7, method: 'doctor', params: { projects: [] } })
    expect(parseRequest('{"jsonrpc":"2.0","method":"note"}').id).toBeNull()
    expect(parseRequest('{"id":null,"method":"note"}').id).toBeNull()
    expect(parseRequest('{"id":2}').method).toBe('')
  })

  it('refuses what is not a request object', () => {
    for (const body of ['', '{', '[]', '"x"', '{"id":"7","method":"a"}', '{"id":1.5,"method":"a"}', '{"id":1,"method":4}']) {
      expect(() => parseRequest(body), body).toThrow(MessageParseError)
    }
  })
})

describe('parseResponse', () => {
  it('reads a result and an error', () => {
    expect(parseResponse('{"jsonrpc":"2.0","id":1,"result":{"a":1}}')).toEqual({ id: 1, result: { a: 1 }, error: null })
    expect(parseResponse('{"id":null,"error":{"code":-32700,"message":"x"}}')).toEqual({ id: null, result: undefined, error: { code: -32700, message: 'x' } })
    expect(() => parseResponse('{"id":1,"error":{"code":"x"}}')).toThrow(MessageParseError)
  })
})
