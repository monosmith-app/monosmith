import { MessageParseError, ParamsError, decodeParams, parseRequest, writeFailure, writeResult } from '../codec/index.ts'
import type { ReceivedRequest } from '../codec/index.ts'
import { workerLimit } from '../concurrency/index.ts'
import { FrameReader, frameMessage } from '../framing/index.ts'
import { ERROR_CODES } from '../protocol/index.ts'
import type { CancelParams } from '../protocol/index.ts'
import { halfImplemented } from './capabilities-of.algorithm.ts'
import { claimStdout } from './claim-stdout.use-case.ts'
import { dispatchRequest } from './dispatch-request.use-case.ts'
import type { SmithPlugin } from './plugin.contract.ts'
import type { FrameSink, ServeStreams } from './protocol-streams.contract.ts'
import { RequestPool } from './request-pool.store.ts'
import type { PooledRequest } from './request-pool.store.ts'

/**
 * Runs a plugin on the process's standard input and output, and exits when the
 * host sends `shutdown` or closes the stream. Use it as the whole body of the
 * plugin's entry point:
 *
 * ```ts
 * await serve({ describe: () => ({ ... }), discover: (params) => ({ ... }) })
 * ```
 *
 * Standard output carries only the protocol's frames: whatever else the plugin
 * or a library writes there goes to standard error instead. A failure of the
 * stream exits with status 1 after the reason is printed.
 */
export async function serve (plugin: SmithPlugin): Promise<never> {
  const output = claimStdout()
  try {
    await serveStreams(plugin, { input: process.stdin, output })
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`)
    // serve is the plugin's main, like the Go SDK's serve.Main: it ends the process, whatever handles the plugin left open.
    // eslint-disable-next-line unicorn/no-process-exit
    process.exit(1)
  }
  // eslint-disable-next-line unicorn/no-process-exit
  process.exit(0)
}

/** What the loop remembers between requests. */
interface Session {
  initialized:          boolean
  readonly plugin:      SmithPlugin
  readonly output:      FrameSink
  readonly pool:        RequestPool
  readonly cancellable: boolean
}

/**
 * Answers requests from `input` on `output` until the host sends `shutdown`
 * or the input ends, then resolves. This is the loop {@link serve} runs, on
 * streams you choose: a test serves a plugin in-process over pipes with it.
 *
 * Semantics of the Go SDK's `serve.Serve`: requests are answered by a bounded
 * pool of as many at a time as the plugin declared in `concurrency` (one, in
 * order, when it said nothing), and the answers may be written in any order;
 * a request before `initialize` is refused (InvalidRequest); a notification
 * (no `id`) gets no answer; an unreadable frame body is answered with
 * ParseError and the loop goes on; an unreadable header ends it with an error.
 * A `cancellable` plugin has the signal of a request aborted when the host
 * sends `cancel` for it, and at shutdown or the end of the input.
 */
export async function serveStreams (plugin: SmithPlugin, streams: ServeStreams): Promise<void> {
  const log = streams.log ?? ((message: string) => process.stderr.write(`${message}\n`))
  for (const member of halfImplemented(plugin)) {
    log(`plugin: ${member} needs its pair (install with uninstall, declared with rewrite) to be advertised; it will not be called`)
  }
  const reader = new FrameReader()
  const pool: RequestPool = new RequestPool(async pooled => { await run(session, pooled) })
  const session: Session = { initialized: false, plugin, output: streams.output, pool, cancellable: plugin.cancellable === true }
  try {
    for await (const chunk of streams.input) {
      for (const body of reader.push(chunk)) {
        if (pool.failure !== undefined) throw pool.failure
        const shutdown = await receive(session, body)
        if (shutdown) return
      }
    }
  } finally {
    await drain(session)
  }
  if (pool.failure !== undefined) throw pool.failure
}

/** Handles one frame. Returns whether it was `shutdown`, after which the loop ends. */
async function receive (session: Session, body: string): Promise<boolean> {
  let request
  try {
    request = parseRequest(body)
  } catch (error) {
    if (!(error instanceof MessageParseError)) throw error
    await send(session.output, writeFailure(null, ERROR_CODES.ParseError, error.message))

    return false
  }
  if (request.id === null) {
    await notified(session, request)

    return false
  }
  const id = request.id
  if (!session.initialized && request.method !== 'initialize') {
    await send(session.output, writeFailure(id, ERROR_CODES.InvalidRequest, 'initialize must come first'))

    return false
  }
  if (request.method === 'initialize') {
    const outcome = await dispatchRequest(session.plugin, request)
    if (outcome.ok) {
      session.initialized = true
      session.pool.setLimit(workerLimit((JSON.parse(outcome.resultJson) as { concurrency?: number }).concurrency))
    }
    await send(session.output, outcome.ok ? writeResult(id, outcome.resultJson) : writeFailure(id, outcome.code, outcome.message))

    return false
  }
  if (request.method === 'shutdown') {
    // Everything read before shutdown, started or still waiting, is answered with its result first, as in the Go SDK; only a cancel message stops a request.
    await session.pool.idle()
    await send(session.output, writeResult(id, '{}'))

    return true
  }
  switch (session.pool.submit({ ...request, id })) {
    case 'duplicate': {
      await send(session.output, writeFailure(id, ERROR_CODES.InvalidRequest, `request id ${id} is already in flight`))
      break
    }
    case 'full': {
      await send(session.output, writeFailure(id, ERROR_CODES.InvalidRequest, 'too many requests in flight: the host must not send more than the plugin\'s declared concurrency'))
      break
    }
    case 'admitted': {
      break
    }
  }

  return false
}

/**
 * Handles a message without an `id`. `cancel` is the only one the protocol
 * defines for a plugin; every other is ignored, and so is a `cancel` to a
 * plugin that is not `cancellable`, or one that names no request.
 */
async function notified (session: Session, request: ReceivedRequest): Promise<void> {
  if (request.method !== 'cancel' || !session.cancellable) return
  let params: CancelParams
  try {
    params = decodeParams('cancel', request.params) as CancelParams
  } catch (error) {
    if (error instanceof ParamsError) return
    throw error
  }
  const waiting = session.pool.cancel(params.id)
  if (waiting !== undefined) await answerCancelled(session.output, waiting)
}

/** A worker's job: handle one request and answer it. */
async function run (session: Session, pooled: PooledRequest): Promise<void> {
  const outcome = await dispatchRequest(session.plugin, pooled.request, pooled.controller.signal)
  const answer = !outcome.ok && outcome.code === ERROR_CODES.PluginFailure && pooled.cancelled
    ? { ok: false as const, code: ERROR_CODES.RequestCancelled, message: `request cancelled: ${outcome.message}` }
    : outcome
  session.pool.release(pooled)
  await send(session.output, answer.ok ? writeResult(pooled.request.id, answer.resultJson) : writeFailure(pooled.request.id, answer.code, answer.message))
}

function answerCancelled (output: FrameSink, pooled: PooledRequest): Promise<void> {
  return send(output, writeFailure(pooled.request.id, ERROR_CODES.RequestCancelled, 'request cancelled before it started'))
}

/**
 * Ends the work in flight when the loop returns (the input ended, the host is gone): a
 * cancellable plugin is asked to stop and what had not started is refused;
 * any other finishes what it is doing.
 */
async function drain (session: Session): Promise<void> {
  if (session.cancellable) {
    for (const waiting of session.pool.cancelAll()) await answerCancelled(session.output, waiting)
  }
  await session.pool.idle()
}

function send (output: FrameSink, body: string): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    output.write(frameMessage(body), (error) => {
      if (error === null || error === undefined) resolve()
      else reject(error)
    })
  })
}
