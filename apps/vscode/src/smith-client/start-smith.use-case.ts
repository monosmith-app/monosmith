import { spawn } from 'node:child_process'
import { openSession } from './json-rpc-session.client.ts'
import { judgeProtocol } from './protocol-range.policy.ts'
import type { SmithInitialized, SmithSession } from './smith-connection.contract.ts'

/**
 * Starts `smith serve` in the workspace and checks what it speaks.
 * @param executable - the smith found by `locateSmith`
 * @param workspaceRoot - the directory to serve
 * @throws when smith cannot start, ends early, or speaks an older protocol
 */
export async function startSmith (executable: string, workspaceRoot: string): Promise<SmithSession> {
  const child = spawn(executable, ['serve'], {
    cwd:   workspaceRoot,
    stdio: ['pipe', 'pipe', 'ignore'],
    shell: executable.toLowerCase().endsWith('.cmd'),
  })
  const connection = openSession(child.stdout, child.stdin)
  child.once('error', (error) => { connection.close(error) })
  const stop = (): void => {
    connection.close()
    child.stdin.end()
    child.kill()
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
    throw error
  }
}
