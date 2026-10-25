import { spawn } from 'node:child_process'
import type { ChildProcess } from 'node:child_process'
import type { StartOptions } from './host-options.contract.ts'
import { PluginClient } from './plugin-client.client.ts'
import type { PluginConnection } from './plugin-client.client.ts'

interface Exit {
  code:   number | null
  signal: NodeJS.Signals | null
}

/**
 * Launches a plugin process and performs the handshake, as smith does: the
 * plugin's argv, `SMITH_PLUGIN=1` in its environment, its standard input and
 * output as the protocol's channel, its standard error forwarded to this
 * process's. Close the client when done: it asks the plugin to shut down and
 * kills it if it does not go.
 */
export async function startPlugin (options: StartOptions): Promise<PluginClient> {
  const [program, ...args] = options.command
  if (program === undefined) throw new Error('plugin command is empty')
  const env = { ...(options.env ?? process.env) }
  env.SMITH_PLUGIN ??= '1'
  const child = spawn(program, args, { env, cwd: options.cwd, stdio: ['pipe', 'pipe', options.stderr ?? 'inherit'], windowsHide: true })
  const exited = new Promise<Exit>((resolve) => {
    child.once('close', (code, signal) => { resolve({ code, signal }) })
  })
  // A program that cannot be started never answers: its streams just end.
  let spawnFailure: Error | undefined
  child.once('error', (cause) => { spawnFailure = cause })
  // Writing to a plugin that has gone raises EPIPE on stdin; the call that wrote sees the error.
  child.stdin.on('error', () => {})

  const connection: PluginConnection = {
    input:  child.stdout,
    output: { write: (chunk, callback) => child.stdin.write(chunk, callback), end: () => child.stdin.end() },
    kill:   () => { child.kill() },
    finish: (graceMs) => waitForExit(child, exited, graceMs),
  }
  try {
    return await PluginClient.open(connection, options)
  } catch (error) {
    child.kill()
    await exited
    if (spawnFailure === undefined) throw error
    throw new Error(`starting plugin ${JSON.stringify(options.command)}: ${spawnFailure.message}`, { cause: error })
  }
}

async function waitForExit (child: ChildProcess, exited: Promise<Exit>, graceMs: number): Promise<string | null> {
  let timer: NodeJS.Timeout | undefined
  const timeout = new Promise<'timeout'>((resolve) => {
    timer = setTimeout(() => {
      resolve('timeout')
    }, graceMs)
  })
  const outcome = await Promise.race([exited, timeout])
  clearTimeout(timer)
  if (outcome === 'timeout') {
    child.kill()
    await exited

    return 'plugin did not exit after shutdown; killed'
  }

  return outcome.code === 0 ? null : `plugin exited with ${outcome.signal ?? `status ${outcome.code}`}`
}
