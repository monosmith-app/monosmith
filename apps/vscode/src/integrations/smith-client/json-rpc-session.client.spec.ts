import { PassThrough } from 'node:stream'
import { describe, expect, it } from 'vitest'
import { openSession } from './json-rpc-session.client.ts'

/** A pretend `smith serve`: what the client wrote comes out as `sent`. */
function pair () {
  const fromServer = new PassThrough()
  const toServer = new PassThrough()
  const sent: Array<Record<string, unknown>> = []
  toServer.on('data', (chunk: Buffer) => {
    const lines = chunk.toString().split('\n').filter(Boolean)
    for (const line of lines) sent.push(JSON.parse(line) as Record<string, unknown>)
  })
  const reply = (message: unknown): void => { fromServer.write(JSON.stringify(message) + '\n') }

  return { connection: openSession(fromServer, toServer), fromServer, sent, reply }
}

describe('openSession', () => {
  it('answers a request with the result that has its id', async () => {
    const { connection, reply, sent } = pair()
    const answer = connection.request<{ ok: boolean }>('ping')
    expect(sent[0]).toMatchObject({ jsonrpc: '2.0', id: 1, method: 'ping' })
    reply({ jsonrpc: '2.0', id: 1, result: { ok: true } })
    await expect(answer).resolves.toEqual({ ok: true })
  })

  it('rejects with the server message when a request failed', async () => {
    const { connection, reply } = pair()
    const answer = connection.request('run')
    reply({ jsonrpc: '2.0', id: 1, error: { code: -32602, message: 'no such project' } })
    await expect(answer).rejects.toThrow('no such project')
  })

  it('hands events to the listeners until they stop', () => {
    const { connection, reply } = pair()
    const seen: unknown[] = []
    const stop = connection.onEvent((event) => { seen.push(event) })
    reply({ jsonrpc: '2.0', method: 'event', params: { event: 'run.started' } })
    stop()
    reply({ jsonrpc: '2.0', method: 'event', params: { event: 'run.finished' } })
    expect(seen).toEqual([{ event: 'run.started' }])
  })

  it('reads a message that arrives in pieces', async () => {
    const { connection, fromServer } = pair()
    const answer = connection.request<number>('ping')
    fromServer.write('{"jsonrpc":"2.0","id":1,')
    fromServer.write('"result":7}\n')
    await expect(answer).resolves.toBe(7)
  })

  it('rejects what is waiting when the server ends', async () => {
    const { connection, fromServer } = pair()
    const answer = connection.request('run')
    fromServer.end()
    await expect(answer).rejects.toThrow('smith serve ended')
    await expect(connection.request('ping')).rejects.toThrow('smith serve ended')
  })
})
