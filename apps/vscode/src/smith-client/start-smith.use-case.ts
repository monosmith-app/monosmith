import { spawn } from 'node:child_process'
import { openSession } from './json-rpc-session.client.ts'
import { judgeProtocol } from './protocol-range.policy.ts'
import { lastLines } from './stderr-tail.algorithm.ts'
import type { SmithExit, SmithInitialized, SmithSession } from './smith-connection.contract.ts'

/** What the caller wants to hear from the process while it lives. */
export interface SmithHooks {
  /** Every line smith writes to stderr, as it arrives. */
  onStderr?: (line: string) => void
  /** The process ended without being stopped by `stop`. */
  onExit?:   (exit: SmithExit) => void
}

/**
 * Starts `smith serve` in the workspace and checks what it speaks.
 * @param executable - the smith found by `locateSmith`
 * @param workspaceRoot - the directory to serve
 * @param hooks - what to hear from the process
 * @throws when smith cannot start, ends early, or speaks an older protocol
 */
export async function startSmith (executable: string, workspaceRoot: string, hooks: SmithHooks = {}): Promise<SmithSession> {
  const shell = executable.toLowerCase().endsWith('.cmd')
  const child = spawn(shell ? `"${executable}"` : executable, ['serve'], {
    cwd:         workspaceRoot,
    stdio:       ['pipe', 'pipe', 'pipe'],
    windowsHide: true,
    shell,
  })
  const connection = openSession(child.stdout, child.stdin)
  let stopped = false
  let tail = ''
  child.stderr.setEncoding('utf8')
  child.stderr.on('data', (chunk: string) => {
    tail = lastLines(tail + chunk, 20)
    for (const line of chunk.split(/\r?\n/)) {
      if (line.trim() !== '') hooks.onStderr?.(line)
    }
  })
  child.once('error', (error) => { connection.close(error) })
  child.once('close', (code, signal) => {
    connection.close(new Error(tail.trim() === '' ? 'smith serve ended' : `smith serve ended: ${lastLines(tail, 3).trim()}`))
    if (!stopped) hooks.onExit?.({ code, signal, stderr: tail.trim() })
  })
  const stop = (): void => {
    stopped = true
    connection.close()
    child.stdin.end()
    // A .cmd shim runs smith as a grandchild; end the whole tree, not just the shell.
    if (shell && child.pid !== undefined) spawn('taskkill', ['/pid', String(child.pid), '/T', '/F'], { stdio: 'ignore', windowsHide: true }).on('error', () => {})
    else child.kill()
  }
  try {
    const hello = await connection.request<SmithInitialized>('initialize')
    const protocol = judgeProtocol(hello.protocolVersion)
    if (protocol === 'older') {
      throw new Error(`smith ${hello.version} speaks protocol ${hello.protocolVersion}, older than this extension supports. Upgrade smith.`)
    }

    return { connection, version: hello.version, protocol: protocol === 'newer' ? 'newer' : 'supported', stop }
  } catch (error) {
    stop()
    const said = lastLines(tail, 3)
    if (said !== '' && error instanceof Error && error.message === 'smith serve ended') {
      throw new Error(`smith serve ended before it answered: ${said}`, { cause: error })
    }
    throw error
  }
}
