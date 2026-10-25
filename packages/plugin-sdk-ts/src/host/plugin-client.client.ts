import { decodeResult, encodeParams, parseResponse, writeNotification, writeRequest } from '../codec/index.ts'
import { inFlightLimit } from '../concurrency/index.ts'
import { FrameReader, frameMessage } from '../framing/index.ts'
import { PROTOCOL_VERSION, RpcError, protocolSchema } from '../protocol/index.ts'
import type { InitializeResult, MethodName, ParamsOf, ResultOf } from '../protocol/index.ts'
import { CallInterruptedError } from './call-interrupted.error.ts'
import type { InterruptionReason } from './call-interrupted.error.ts'
import { DEFAULT_CALL_TIMEOUT_MS, DEFAULT_CANCEL_GRACE_MS } from './host-options.contract.ts'
import type { CallOptions, HostOptions } from './host-options.contract.ts'

/** The two ends of the pipe to a plugin, and what to do with the plugin when the stream cannot be trusted. */
export interface PluginConnection {
  /** The plugin's responses. */
  input:  AsyncIterable<Uint8Array>
  /** The plugin's requests. */
  output: { write: (chunk: Buffer, callback: (error?: Error | null) => void) => unknown, end: () => unknown }
  /** Called when the plugin did not answer in time: ends the plugin, if it is a process. */
  kill:   () => void
  /** Resolves once the plugin has gone, with the reason it did not end well (or `null`). Called by {@link PluginClient.close}. */
  finish: (graceMs: number) => Promise<string | null>
}

/** How a call ends without a response of its own: the plugin went, or did not answer. */
type Outcome = { body: string } | { error: Error }

interface Waiter {
  settle: (outcome: Outcome) => void
}

/** A call waiting for a place among the requests that may be in flight. */
interface Queued {
  grant: () => void
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
 * author's tests (smith itself is the host in real use). It never has more
 * requests in flight than the plugin declared in `concurrency`, as smith does:
 * with the default of one, calls are serialised, each waiting for the one
 * before it to be answered. Each call has a timeout and may be cancelled with
 * an `AbortSignal`. Open one with {@link startPlugin} (a process),
 * {@link connectPlugin} (streams) or {@link connectInProcess} (a plugin object
 * served in this process).
 */
export class PluginClient {
  /** Performs the handshake on a connection. */
  static async open (connection: PluginConnection, options: HostOptions): Promise<PluginClient> {
    const client = new PluginClient(connection, options.cancelGraceMs ?? DEFAULT_CANCEL_GRACE_MS)
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
    client.#limit = inFlightLimit(client.info.concurrency)

    return client
  }

  readonly #connection: PluginConnection
  #waiters = new Map<number, Waiter>()
  readonly #queue:      Queued[] = []
  readonly #graceMs:    number
  #lost:                Error | undefined
  #nextId = 0
  #limit = 1
  #inFlight = 0
  #peak = 0
  /** What the plugin answered to `initialize`: identity, kind, roles, capabilities. */
  info!:                InitializeResult

  private constructor (connection: PluginConnection, graceMs: number) {
    this.#connection = connection
    this.#graceMs = graceMs
    void this.#readResponses()
  }

  /** Waits for a place among the requests that may be in flight; rejects if `signal` aborts first. */
  #acquire (signal: AbortSignal | undefined): Promise<void> {
    if (signal?.aborted === true) return Promise.reject(new CallInterruptedError('cancelled before it was sent', 'cancelled', false))
    if (this.#inFlight < this.#limit) {
      this.#take()

      return Promise.resolve()
    }

    return new Promise<void>((resolve, reject) => {
      const entry: Queued = {
        grant: () => {
          signal?.removeEventListener('abort', onAbort)
          this.#take()
          resolve()
        },
      }
      const onAbort = (): void => {
        const at = this.#queue.indexOf(entry)
        if (at !== -1) this.#queue.splice(at, 1)
        reject(new CallInterruptedError('cancelled before it was sent', 'cancelled', false))
      }
      signal?.addEventListener('abort', onAbort, { once: true })
      this.#queue.push(entry)
    })
  }

  #take (): void {
    this.#inFlight++
    this.#peak = Math.max(this.#peak, this.#inFlight)
  }

  #release (): void {
    this.#inFlight--
    this.#queue.shift()?.grant()
  }

  async #send (method: string, params: unknown, timeoutMs: number, signal: AbortSignal | undefined): Promise<unknown> {
    if (this.#lost !== undefined) throw new Error(`${method}: plugin stopped answering: ${this.#lost.message}`, { cause: this.#lost })
    const id = ++this.#nextId
    const body = writeRequest(id, method, params === undefined ? undefined : encodeParams(method, params))
    const answer = new Promise<Outcome>((resolve) => { this.#waiters.set(id, { settle: resolve }) })
    try {
      await writeFrame(this.#connection.output, frameMessage(body))
    } catch (error) {
      this.#waiters.delete(id)
      throw new Error(`${method}: plugin is not accepting requests: ${messageOf(error)}`, { cause: error })
    }
    let timer: NodeJS.Timeout | undefined
    let onAbort: (() => void) | undefined
    const interrupted = new Promise<InterruptionReason>((resolve) => {
      timer = setTimeout(() => { resolve('timeout') }, timeoutMs)
      onAbort = () => { resolve('cancelled') }
      if (signal?.aborted === true) resolve('cancelled')
      else signal?.addEventListener('abort', onAbort, { once: true })
    })
    try {
      const winner = await Promise.race([answer, interrupted])
      if (typeof winner !== 'string') {
        this.#waiters.delete(id)

        return this.#decode(method, winner)
      }

      return await this.#giveUp(method, id, answer, winner, timeoutMs)
    } finally {
      clearTimeout(timer)
      if (onAbort !== undefined) signal?.removeEventListener('abort', onAbort)
    }
  }

  #decode (method: string, outcome: Outcome): unknown {
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

  /**
   * The end of a call whose signal aborted or whose time ran out. A plugin that
   * advertised cancel is sent the notification and has the grace period to
   * answer; one that cannot be asked is killed at once when the time ran out,
   * and after the grace period when the caller cancelled. A plugin that does not
   * answer in time is killed, and every other call in flight to it fails.
   */
  async #giveUp (method: string, id: number, answer: Promise<Outcome>, reason: InterruptionReason, timeoutMs: number): Promise<unknown> {
    const asked = this.supports('cancel')
    if (asked) {
      try {
        await this.notify('cancel', { id })
      } catch {
        // The plugin is gone; waiting for the answer below reports how it went.
      }
    }
    if (asked || reason === 'cancelled') {
      let timer: NodeJS.Timeout | undefined
      const grace = new Promise<'expired'>((resolve) => {
        timer = setTimeout(() => { resolve('expired') }, this.#graceMs)
      })
      const winner = await Promise.race([answer, grace])
      clearTimeout(timer)
      if (winner !== 'expired') {
        this.#waiters.delete(id)

        return this.#answeredAfterStopping(method, winner, reason, timeoutMs)
      }
    }
    this.#waiters.delete(id)
    this.#connection.kill()
    if (reason === 'timeout') throw new CallInterruptedError(`${method}: plugin did not answer within ${timeoutMs}ms`, reason, true)
    throw new CallInterruptedError(`${method}: cancelled; the plugin did not stop within ${this.#graceMs}ms and was killed`, reason, true)
  }

  /** The plugin answered while the host was giving up: a result is the work finished in time, and is the call's result. */
  #answeredAfterStopping (method: string, outcome: Outcome, reason: InterruptionReason, timeoutMs: number): unknown {
    try {
      return this.#decode(method, outcome)
    } catch (error) {
      const what = reason === 'timeout' ? `did not answer within ${timeoutMs}ms` : 'cancelled'
      throw new CallInterruptedError(`${method}: ${what}: ${messageOf(error)}`, reason, false, error instanceof RpcError ? error : undefined)
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
    let id: number | null
    try {
      id = parseResponse(body).id
    } catch (error) {
      // A malformed response cannot be attributed to a call: the stream is lost.
      this.#lose(new Error(`malformed response: ${messageOf(error)}`, { cause: error }))

      return
    }
    if (id === null) return
    const waiter = this.#waiters.get(id)
    if (waiter === undefined) return
    this.#waiters.delete(id)
    waiter.settle({ body })
  }

  #lose (error: Error): void {
    this.#lost = error
    const waiting = this.#waiters
    this.#waiters = new Map()
    for (const waiter of waiting.values()) waiter.settle({ error })
  }

  /**
   * Sends one request and resolves with its decoded result. An error answer
   * rejects with an {@link RpcError} carrying its code. The call waits for a
   * place first, so the plugin never has more requests in flight than it
   * declared; the timeout counts from the moment the request is sent.
   *
   * When `options.signal` aborts or the time runs out the host stops waiting.
   * A plugin that advertised `cancel` is sent the notification and has the
   * grace period to answer: with the result it had already (the call resolves),
   * or an error (the call rejects with a {@link CallInterruptedError} carrying
   * it). A plugin that does not answer in time is killed, because its stream can
   * no longer be trusted, and the call rejects; a plugin that cannot be asked
   * to stop is killed at once on a timeout, and after the grace period when
   * the caller cancelled.
   */
  call<M extends MethodName>(method: M, params: ParamsOf<M>, options?: CallOptions): Promise<ResultOf<M>>
  call (method: string, params?: unknown, options?: CallOptions): Promise<unknown>
  async call (method: string, params?: unknown, options: CallOptions = {}): Promise<unknown> {
    const timeoutMs = options.timeoutMs ?? DEFAULT_CALL_TIMEOUT_MS
    if (method === 'shutdown') return await this.#send(method, params, timeoutMs, options.signal) // it must reach a plugin that is busy
    try {
      await this.#acquire(options.signal)
    } catch (error) {
      if (error instanceof CallInterruptedError) throw new CallInterruptedError(`${method}: ${error.message}`, error.reason, false)
      throw error
    }
    try {
      return await this.#send(method, params, timeoutMs, options.signal)
    } finally {
      this.#release()
    }
  }

  /** How long the plugin has to answer after it was asked to stop. */
  get cancelGraceMs (): number {
    return this.#graceMs
  }

  /** The most requests this client has had in flight at once. */
  get peakInFlight (): number {
    return this.#peak
  }

  /** Sends a message that gets no answer, such as `cancel`. */
  async notify (method: string, params: unknown): Promise<void> {
    if (this.#lost !== undefined) throw new Error(`${method}: plugin stopped answering: ${this.#lost.message}`, { cause: this.#lost })
    try {
      const body = writeNotification(method, encodeParams(method, params))
      await writeFrame(this.#connection.output, frameMessage(body))
    } catch (error) {
      throw new Error(`${method}: plugin is not accepting requests: ${messageOf(error)}`, { cause: error })
    }
  }

  /** Whether the plugin advertised a method: `initialize` and `shutdown` always, the others by their capability. */
  supports (method: string): boolean {
    if (method === 'initialize' || method === 'shutdown') return true
    const capability = Object.hasOwn(protocolSchema['x-methods'], method) ? protocolSchema['x-methods'][method].capability : null

    const advertised = this.info?.capabilities as Record<string, boolean | undefined> | undefined

    return capability !== null && advertised?.[capability] === true
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
