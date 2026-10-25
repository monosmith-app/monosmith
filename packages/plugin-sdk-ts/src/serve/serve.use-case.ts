import { MessageParseError, parseRequest, writeFailure, writeResult } from '../codec/index.ts'
import { FrameReader, frameMessage } from '../framing/index.ts'
import { ERROR_CODES } from '../protocol/index.ts'
import { halfImplemented } from './capabilities-of.algorithm.ts'
import { claimStdout } from './claim-stdout.use-case.ts'
import { dispatchRequest } from './dispatch-request.use-case.ts'
import type { SmithPlugin } from './plugin.contract.ts'
import type { FrameSink, ServeStreams } from './protocol-streams.contract.ts'

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
  initialized: boolean
}

/**
 * Answers requests from `input` on `output` until the host sends `shutdown`
 * or the input ends, then resolves. This is the loop {@link serve} runs, on
 * streams you choose: a test serves a plugin in-process over pipes with it.
 *
 * Semantics of the Go SDK's `serve.Serve`: requests are answered one at a time,
 * in order; a request before `initialize` is refused (InvalidRequest);
 * a notification (no `id`) gets no answer; an unreadable frame body is answered
 * with ParseError and the loop goes on; an unreadable header ends it with an error.
 */
export async function serveStreams (plugin: SmithPlugin, streams: ServeStreams): Promise<void> {
  const log = streams.log ?? ((message: string) => process.stderr.write(`${message}\n`))
  for (const member of halfImplemented(plugin)) {
    log(`plugin: ${member} needs its pair (install with uninstall, declared with rewrite) to be advertised; it will not be called`)
  }
  const reader = new FrameReader()
  const session: Session = { initialized: false }
  for await (const chunk of streams.input) {
    for (const body of reader.push(chunk)) {
      const shutdown = await answer(plugin, streams.output, session, body)
      if (shutdown) return
    }
  }
}

/** Answers one frame. Returns whether it was `shutdown`, after which the loop ends. */
async function answer (plugin: SmithPlugin, output: FrameSink, session: Session, body: string): Promise<boolean> {
  let request
  try {
    request = parseRequest(body)
  } catch (error) {
    if (!(error instanceof MessageParseError)) throw error
    await send(output, writeFailure(null, ERROR_CODES.ParseError, error.message))

    return false
  }
  // A notification has no reply; version 1 defines none for plugins.
  if (request.id === null) return false
  if (!session.initialized && request.method !== 'initialize') {
    await send(output, writeFailure(request.id, ERROR_CODES.InvalidRequest, 'initialize must come first'))

    return false
  }
  const outcome = await dispatchRequest(plugin, request)
  if (request.method === 'initialize' && outcome.ok) session.initialized = true
  await send(output, outcome.ok ? writeResult(request.id, outcome.resultJson) : writeFailure(request.id, outcome.code, outcome.message))

  return request.method === 'shutdown'
}

function send (output: FrameSink, body: string): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    output.write(frameMessage(body), (error) => {
      if (error === null || error === undefined) resolve()
      else reject(error)
    })
  })
}
