import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { pathToFileURL } from 'node:url'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { startPlugin } from './start-plugin.use-case.ts'

const sdk = pathToFileURL(join(import.meta.dirname, '..', 'index.ts')).href
let dir: string

beforeEach(() => { dir = mkdtempSync(join(tmpdir(), 'sdk-start-')) })
afterEach(() => { rmSync(dir, { recursive: true, force: true }) })

function script (body: string): string {
  const file = join(dir, 'plugin.ts')
  writeFileSync(file, `import { serve } from ${JSON.stringify(sdk)}\n${body}\n`)

  return file
}

describe('startPlugin', () => {
  it('launches the process with SMITH_PLUGIN=1, shakes hands, calls it and shuts it down', async () => {
    const file = script(`
      await serve({
        describe: () => ({ plugin: { name: 'proc', version: '1.0.0' }, kind: 'convention', languages: [process.env.SMITH_PLUGIN ?? 'unset'] }),
        doctor: () => { console.log('a stray line on stdout'); return { findings: [{ check: 'c', ok: true }] } },
      })`)
    const client = await startPlugin({ command: [process.execPath, file], workspaceRoot: dir, stderr: 'ignore' })
    expect(client.info.languages).toEqual(['1'])
    expect(client.info.capabilities).toEqual({ doctor: true })
    // The stray console.log would corrupt the stream if it reached stdout.
    await expect(client.call('doctor', { projects: [] })).resolves.toEqual({ findings: [{ check: 'c', ok: true }] })
    await client.close()
  })

  it('kills a plugin that does not answer in time', async () => {
    const file = script('await serve({ describe: () => ({ plugin: { name: \'p\', version: \'1\' }, kind: \'convention\' }), doctor: () => new Promise(() => setInterval(() => {}, 1000)) })')
    const client = await startPlugin({ command: [process.execPath, file], workspaceRoot: dir, stderr: 'ignore' })
    await expect(client.call('doctor', { projects: [] }, { timeoutMs: 300 })).rejects.toThrow('did not answer within 300ms')
    await expect(client.close()).rejects.toThrow()
  })

  it('names the command when the program does not exist', async () => {
    await expect(startPlugin({ command: [join(dir, 'no-such-program')], workspaceRoot: dir, stderr: 'ignore' })).rejects.toThrow('starting plugin')
  })

  it('fails the handshake of a process that exits at once', async () => {
    const file = join(dir, 'exits.mjs')
    writeFileSync(file, 'process.exit(3)\n')
    await expect(startPlugin({ command: [process.execPath, file], workspaceRoot: dir, stderr: 'ignore' })).rejects.toThrow('initializing plugin')
  })

  it('refuses an empty command', async () => {
    await expect(startPlugin({ command: [], workspaceRoot: dir })).rejects.toThrow('plugin command is empty')
  })
})
