import type { Readable, Writable } from 'node:stream'
import type { SmithConnection } from './smith-connection.contract.ts'

interface Message {
  id?:     number
  result?: unknown
  error?:  { message?: string }
  method?: string
  params?: unknown
}

interface Waiting {
  resolve: (value: unknown) => void
  reject:  (reason: Error) => void
}

/**
 * Speaks JSON-RPC 2.0 to `smith serve` over its stdin and stdout, one JSON
 * message per line.
 * @param input - what the server writes
 * @param output - what the server reads
 */
export function openSession (input: Readable, output: Writable): SmithConnection {
  let nextId = 1
  let buffered = ''
  let closed: Error | undefined
  const waiting = new Map<number, Waiting>()
  const listeners = new Set<(event: unknown) => void>()

  const close = (reason: Error = new Error('smith serve ended')): void => {
    closed ??= reason
    for (const call of waiting.values()) call.reject(closed)
    waiting.clear()
  }

  const receive = (line: string): void => {
    if (line.trim() === '') return
    let message: Message
    try {
      message = JSON.parse(line) as Message
    } catch {
      return
    }
    const call = message.id === undefined ? undefined : waiting.get(message.id)
    if (call !== undefined && message.id !== undefined) {
      waiting.delete(message.id)
      if (message.error === undefined) call.resolve(message.result)
      else call.reject(new Error(message.error.message ?? 'smith serve failed the request'))
    } else if (message.method === 'event') {
      for (const listener of listeners) listener(message.params)
    }
  }

  input.setEncoding('utf8')
  input.on('data', (chunk: string) => {
    buffered += chunk
    for (let end = buffered.indexOf('\n'); end >= 0; end = buffered.indexOf('\n')) {
      receive(buffered.slice(0, end))
      buffered = buffered.slice(end + 1)
    }
  })
  input.on('close', () => { close() })

  return {
    request: async <T>(method: string, params?: unknown): Promise<T> => {
      if (closed !== undefined) throw closed
      const id = nextId++
      const answer = new Promise<unknown>((resolve, reject) => { waiting.set(id, { resolve, reject }) })
      output.write(JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n')

      return await (answer as Promise<T>)
    },
    onEvent: (handler) => {
      listeners.add(handler)

      return () => { listeners.delete(handler) }
    },
    close,
  }
}
