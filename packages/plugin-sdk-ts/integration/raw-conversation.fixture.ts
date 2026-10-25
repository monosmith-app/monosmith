import { spawn, spawnSync } from 'node:child_process'
import { FrameReader, frameMessage } from '../src/framing/index.ts'

/** A request to send as it is: the id is the position in the conversation, from 1. */
export interface RawRequest {
  method:  string
  params?: unknown
}

/** Whether a program runs: `smith version`-like probes for tests that need a tool on the path. */
export function isAvailable (program: string, args: string[] = ['--version']): boolean {
  return spawnSync(program, args, { stdio: 'ignore' }).status === 0
}

/**
 * Starts a plugin the way smith does (`SMITH_PLUGIN=1`), writes every request
 * as a frame, and returns the body of every frame it answered with, exactly as
 * the plugin wrote it. The conversation ends when the plugin answers
 * `shutdown`, which this adds to the end.
 */
export async function converse (command: string[], cwd: string, requests: RawRequest[]): Promise<string[]> {
  const [program = '', ...args] = command
  const child = spawn(program, args, { cwd, env: { ...process.env, SMITH_PLUGIN: '1' }, stdio: ['pipe', 'pipe', 'inherit'] })
  const reader = new FrameReader()
  const bodies: string[] = []
  const closed = new Promise<void>((resolve, reject) => {
    child.on('error', reject)
    child.stdout.on('data', (chunk: Buffer) => {
      bodies.push(...reader.push(chunk))
    })
    child.on('close', () => {
      resolve()
    })
  })
  const all = [...requests, { method: 'shutdown', params: {} }]
  for (const [i, request] of all.entries()) child.stdin.write(frameMessage(JSON.stringify({ jsonrpc: '2.0', id: i + 1, ...request })))
  // A plugin that refused every request (none came after initialize) does not stop at shutdown: closing its input does.
  child.stdin.end()
  await closed

  return bodies
}
