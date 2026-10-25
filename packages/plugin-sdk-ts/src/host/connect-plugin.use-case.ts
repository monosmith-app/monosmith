import { PassThrough } from 'node:stream'
import type { SmithPlugin } from '../serve/index.ts'
import { serveStreams } from '../serve/index.ts'
import type { HostOptions } from './host-options.contract.ts'
import { PluginClient } from './plugin-client.client.ts'

/** The streams of a plugin that is somewhere else: its responses to read, its requests to write. */
export interface PluginStreams {
  /** What the plugin writes (its standard output). */
  input:  AsyncIterable<Uint8Array>
  /** What the plugin reads (its standard input). */
  output: { write: (chunk: Buffer, callback: (error?: Error | null) => void) => unknown, end: () => unknown }
}

/**
 * Performs the handshake with a plugin reachable through a pair of streams.
 * Closing the client shuts the plugin down and ends the streams; there is no
 * process to kill.
 */
export function connectPlugin (streams: PluginStreams, options: HostOptions): Promise<PluginClient> {
  return PluginClient.open({
    input:  streams.input,
    output: streams.output,
    kill:   () => { streams.output.end() },
    finish: () => Promise.resolve(null),
  }, options)
}

/**
 * Serves a plugin object in this process over pipes and connects to it, the way
 * the Go SDK's tests serve a plugin with `serve.Serve` over `io.Pipe`: the
 * real framing, handshake and dispatch, no process. Closing the client waits
 * for the serve loop to end.
 */
export async function connectInProcess (plugin: SmithPlugin, options: HostOptions): Promise<PluginClient> {
  const toPlugin = new PassThrough()
  const fromPlugin = new PassThrough()
  const served = (async (): Promise<string | null> => {
    try {
      await serveStreams(plugin, { input: toPlugin, output: fromPlugin, log: () => {} })

      return null
    } catch (error) {
      return error instanceof Error ? error.message : String(error)
    } finally {
      fromPlugin.end()
    }
  })()

  return await PluginClient.open({
    input:  fromPlugin,
    output: toPlugin,
    kill:   () => { toPlugin.end() },
    finish: () => served,
  }, options)
}
