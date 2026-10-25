import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { afterEach, beforeEach, test } from 'node:test'
import { checkPlugin, connectInProcess } from '@monosmith/plugin-sdk'
import { createLayoutPlugin, findJunkDrawers } from './layout.ts'

let workspace = ''
beforeEach(() => { workspace = mkdtempSync(join(tmpdir(), 'layout-')) })
afterEach(() => { rmSync(workspace, { recursive: true, force: true }) })

// The plugin is served in this process over pipes, with the real handshake: the way smith talks to it.
const connect = () => connectInProcess(createLayoutPlugin(), { workspaceRoot: workspace })

test('the plugin conforms', async () => {
  const client = await connect()
  const report = await checkPlugin(client)
  assert.ok(report.ok, report.failures())
  await client.close()
})

test('the handshake advertises exactly what is implemented', async () => {
  const client = await connect()
  assert.deepEqual(client.info.roles, ['structure'])
  assert.deepEqual(client.info.capabilities, { contribute: true, doctor: true })
  await client.close()
})

test('lint waits for the layout check', async () => {
  const client = await connect()
  const result = await client.call('contribute', { projects: [{ name: 'core', root: 'libs/core', language: 'widget', kind: 'lib' }] })
  const core = result.projects.core
  assert.ok(core?.tasks?.['layout-check'], JSON.stringify(core))
  assert.deepEqual(core.dependsOn?.lint, ['layout-check'])
  await client.close()
})

test('an unknown method is -32601', async () => {
  const client = await connect()
  await assert.rejects(client.call('publish', { project: { name: 'a', root: 'a', language: 'widget', kind: 'lib' }, version: '1.0.0' }), { code: -32601 })
  await client.close()
})

test('doctor warns about a missing README', async () => {
  mkdirSync(join(workspace, 'libs', 'core'), { recursive: true })
  const client = await connect()
  const result = await client.call('doctor', { projects: [{ name: 'core', root: 'libs/core', language: 'widget', kind: 'lib' }] })
  const finding = result.findings[0]
  assert.equal(finding?.ok, false)
  assert.equal(finding.warning, true)
  assert.ok(finding.remedy)
  await client.close()
})

test('findJunkDrawers skips build output', () => {
  for (const folder of ['src/orders', 'src/Utils', 'dist/helpers', 'common']) mkdirSync(join(workspace, ...folder.split('/')), { recursive: true })
  assert.deepEqual(findJunkDrawers(workspace).sort((a, b) => a.localeCompare(b)), ['common', 'src/Utils'])
})
