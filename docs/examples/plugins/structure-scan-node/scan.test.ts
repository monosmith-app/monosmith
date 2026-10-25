import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { afterEach, beforeEach, test } from 'node:test'
import { CallInterruptedError, checkPlugin, connectInProcess } from '@monosmith/plugin-sdk'
import type { Project } from '@monosmith/plugin-sdk'
import { createScanPlugin } from './scan.ts'

let workspace = ''
beforeEach(() => { workspace = mkdtempSync(join(tmpdir(), 'scan-')) })
afterEach(() => { rmSync(workspace, { recursive: true, force: true }) })

/** Writes files (name to size in bytes) under libs/core and returns the project. */
function core (files: Record<string, number>): Project {
  for (const [name, size] of Object.entries(files)) {
    const path = join(workspace, 'libs', 'core', ...name.split('/'))
    mkdirSync(dirname(path), { recursive: true })
    writeFileSync(path, Buffer.alloc(size))
  }

  return { name: 'core', root: 'libs/core', language: 'widget', kind: 'lib' }
}

const connect = (delayMs = 0) => connectInProcess(createScanPlugin(delayMs), { workspaceRoot: workspace })

test('the plugin conforms', async () => {
  const client = await connect()
  const report = await checkPlugin(client)
  assert.ok(report.ok, report.failures())
  await client.close()
})

test('it advertises cancel and a concurrency of four', async () => {
  const client = await connect()
  assert.equal(client.info.capabilities.cancel, true)
  assert.equal(client.info.concurrency, 4)
  await client.close()
})

test('doctor warns about a big file', async () => {
  const project = core({ 'src/small.txt': 10, 'data/huge.bin': (1 << 20) + 1 })
  const client = await connect()
  const { findings } = await client.call('doctor', { projects: [project] })
  assert.equal(findings[0]?.ok, false)
  assert.equal(findings[0].warning, true)
  assert.equal(findings[0].detail, 'data/huge.bin')
  await client.close()
})

test('a cancelled scan stops midway and the answer is the cancelled error', async () => {
  const project = core(Object.fromEntries(['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'].map(name => [`${name}.txt`, 1])))
  const client = await connect(100) // eight files: the scan would take 800ms
  const controller = new AbortController()
  const began = Date.now()
  const call = client.call('doctor', { projects: [project] }, { signal: controller.signal })
  setTimeout(() => { controller.abort() }, 150)
  await assert.rejects(call, (error: unknown) => error instanceof CallInterruptedError && error.reason === 'cancelled' && !error.killed && error.response?.code === -32800)
  assert.ok(Date.now() - began < 600, 'the scan went on after being cancelled')
  await client.close()
})

test('four scans run at the same time', async () => {
  const project = core({ 'a.txt': 1, 'b.txt': 1, 'c.txt': 1 })
  const client = await connect(100)
  const began = Date.now()
  await Promise.all(Array.from({ length: 4 }, () => client.call('doctor', { projects: [project] })))
  // Three files at 100ms each is 300ms a scan: one after the other they take 1.2s.
  assert.ok(Date.now() - began < 900)
  assert.equal(client.peakInFlight, 4)
  await client.close()
})
