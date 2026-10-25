import { decodeResult, encodeParams, parseResponse, writeRequest } from '../codec/index.ts'
import { FrameReader, frameMessage } from '../framing/index.ts'
import { PROTOCOL_VERSION, RpcError, protocolSchema } from '../protocol/index.ts'
import type { InitializeResult, MethodName, ParamsOf, ResultOf } from '../protocol/index.ts'
import { DEFAULT_CALL_TIMEOUT_MS } from './host-options.contract.ts'
import type { CallOptions, HostOptions } from './host-options.contract.ts'

/** The two ends of the pipe to a plugin, and what to do with the plugin when the stream cannot be trusted. */
export interface PluginConnection {
  /** The plugin's responses. */
  input:  AsyncIterable<Uint8Array>
  /** The plugin's requests. */
  output: { write: (chunk: Buffer, callback: (error?: Error | null) => void) => unknown, end: () => unknown }
  /** Called when a call timed out: ends the plugin, if it is a process. */
  kill:   () => void
  /** Resolves once the plugin has gone, with the reason it did not end well (or `null`). Called by {@link PluginClient.close}. */
  finish: (graceMs: number) => Promise<string | null>
}

/** How a call ends without a response of its own: the plugin went, or did not answer. */
type Outcome = { body: string } | { error: Error } | { timedOut: true }

interface Waiter {
  id:     number
  settle: (outcome: Outcome) => void
}

function writeFrame (output: PluginConnection['output'], frame: Buffer): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    output.write(frame, (error) => {
      if (error === null || error === undefined) resolve()
      else reject(error)
    })
  })
}

function messageOf (thrown: unknown): string {
  return thrown instanceof Error ? thrown.message : String(thrown)
}

/**
 * The host's side of a connection to one initialized plugin, for a plugin
 * author's tests (smith itself is the host in real use). Calls are serialised,
 * as smith serialises them, each with a timeout. Open one with {@link startPlugin}
 * (a process), {@link connectPlugin} (streams) or {@link connectInProcess}
 * (a plugin object served in this process).
 */
export class PluginClient {
  /** Performs the handshake on a connection. */
  static async open (connection: PluginConnection, options: HostOptions): Promise<PluginClient> {
    const client = new PluginClient(connection)
    try {
      client.info = await client.call('initialize', {
        protocolVersions: [PROTOCOL_VERSION],
        host:             options.host ?? { name: 'smith', version: 'dev' },
        workspaceRoot:    options.workspaceRoot,
      }, { timeoutMs: options.initializeTimeoutMs ?? 30_000 })
    } catch (error) {
      connection.kill()
      throw new Error(`initializing plugin: ${messageOf(error)}`, { cause: error })
    }
    if (client.info.protocolVersion !== PROTOCOL_VERSION) {
      connection.kill()
      throw new Error(`plugin ${client.info.plugin.name} chose protocol version ${client.info.protocolVersion}, host speaks ${PROTOCOL_VERSION}`)
    }

    return client
  }

  #connection: PluginConnection
  #waiter:     Waiter | undefined
  #lost:       Error | undefined
  #tail:       Promise<void> = Promise.resolve()
  #nextId = 0
  /** What the plugin answered to `initialize`: identity, kind, roles, capabilities. */
  info!:       InitializeResult

  private constructor (connection: PluginConnection) {
    this.#connection = connection
    void this.#readResponses()
  }

  /** Runs a task after the ones before it, whatever became of them. */
  async #serialise<T> (task: () => Promise<T>): Promise<T> {
    const previous = this.#tail
    const current = (async (): Promise<T> => {
      await previous

      return await task()
    })()
    this.#tail = (async (): Promise<void> => {
      try {
        await current
      } catch {
        // The caller of this task sees the failure; the next one only waits.
      }
    })()

    return await current
  }

  async #send (method: string, params: unknown, timeoutMs: number): Promise<unknown> {
    if (this.#lost !== undefined) throw new Error(`${method}: plugin stopped answering: ${this.#lost.message}`, { cause: this.#lost })
    const id = ++this.#nextId
    const body = writeRequest(id, method, params === undefined ? undefined : encodeParams(method, params))
    const answer = new Promise<Outcome>((resolve) => { this.#waiter = { id, settle: resolve } })
    try {
      await writeFrame(this.#connection.output, frameMessage(body))
    } catch (error) {
      this.#waiter = undefined
      throw new Error(`${method}: plugin is not accepting requests: ${messageOf(error)}`, { cause: error })
    }
    let timer: NodeJS.Timeout | undefined
    const timeout = new Promise<Outcome>((resolve) => {
      timer = setTimeout(() => {
        this.#connection.kill()
        resolve({ timedOut: true })
      }, timeoutMs)
    })
    const outcome = await Promise.race([answer, timeout])
    clearTimeout(timer)
    this.#waiter = undefined

    if ('timedOut' in outcome) throw new Error(`${method}: plugin did not answer within ${timeoutMs}ms`)
    if ('error' in outcome) throw new Error(`${method}: plugin stopped answering: ${outcome.error.message}`, { cause: outcome.error })
    try {
      const response = parseResponse(outcome.body)
      if (response.error !== null) throw new RpcError(response.error.code, response.error.message)

      return decodeResult(method, response.result)
    } catch (error) {
      if (error instanceof RpcError) throw error
      throw new Error(`${method}: plugin stopped answering: ${messageOf(error)}`, { cause: error })
    }
  }

  async #readResponses (): Promise<void> {
    const reader = new FrameReader()
    try {
      for await (const chunk of this.#connection.input) {
        for (const body of reader.push(chunk)) this.#deliver(body)
      }
      this.#lose(new Error('EOF'))
    } catch (error) {
      this.#lose(error instanceof Error ? error : new Error(String(error)))
    }
  }

  #deliver (body: string): void {
    const waiter = this.#waiter
    if (waiter === undefined) return
    try {
      if (parseResponse(body).id === waiter.id) waiter.settle({ body })
      // Responses to other ids cannot happen with serialised calls; skip them.
    } catch (error) {
      waiter.settle({ error: new Error(`malformed response: ${messageOf(error)}`, { cause: error }) })
    }
  }

  #lose (error: Error): void {
    this.#lost = error
    this.#waiter?.settle({ error })
  }

  /**
   * Sends one request and resolves with its decoded result. An error answer
   * rejects with an {@link RpcError} carrying its code. If the plugin does not
   * answer in time, the plugin is killed and the call rejects.
   */
  call<M extends MethodName>(method: M, params: ParamsOf<M>, options?: CallOptions): Promise<ResultOf<M>>
  call (method: string, params?: unknown, options?: CallOptions): Promise<unknown>
  async call (method: string, params?: unknown, options: CallOptions = {}): Promise<unknown> {
    return await this.#serialise(() => this.#send(method, params, options.timeoutMs ?? DEFAULT_CALL_TIMEOUT_MS))
  }

  /** Whether the plugin advertised a method: `initialize` and `shutdown` always, the others by their capability. */
  supports (method: string): boolean {
    if (method === 'initialize' || method === 'shutdown') return true
    const capability = Object.hasOwn(protocolSchema['x-methods'], method) ? protocolSchema['x-methods'][method].capability : null

    return capability !== null && (this.info.capabilities as Record<string, boolean | undefined>)[capability] === true
  }

  /** Asks the plugin to shut down and waits briefly for it to go. Rejects if it did not go quietly. */
  async close (): Promise<void> {
    try {
      await this.call('shutdown', {}, { timeoutMs: 5000 })
    } catch {
      // A plugin that is already gone cannot be asked; waiting for it below reports how it went.
    }
    this.#connection.output.end()
    const trouble = await this.#connection.finish(5000)
    if (trouble !== null) throw new Error(trouble)
  }
}
