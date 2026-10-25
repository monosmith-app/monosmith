import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { afterEach, beforeEach, test } from 'node:test'
import { checkPlugin, connectInProcess } from '@monosmith/plugin-sdk'
import { createWidgetPlugin } from './widget.ts'

let workspace = ''
beforeEach(() => { workspace = mkdtempSync(join(tmpdir(), 'widget-')) })
afterEach(() => { rmSync(workspace, { recursive: true, force: true }) })

// The plugin is served in this process over pipes, with the real handshake: the way smith talks to it.
const connect = () => connectInProcess(createWidgetPlugin(), { workspaceRoot: workspace })

function write (relative: string, content: string): void {
  const full = join(workspace, ...relative.split('/'))
  mkdirSync(dirname(full), { recursive: true })
  writeFileSync(full, content)
}

const core = { name: 'core', root: 'libs/core', language: 'widget', kind: 'lib' as const }

// The conformance suite is what `smith plugin check --generate` runs.
test('the plugin conforms, generating an app, a lib and a package that it discovers again', async () => {
  const client = await connect()
  const report = await checkPlugin(client, { workspace })
  assert.ok(report.ok, report.failures())
  await client.close()
})

test('discover turns uses into imports', async () => {
  write('libs/core/widget.json', '{"name":"core","version":"1.0.0"}')
  write('apps/shop/widget.json', '{"name":"shop","version":"1.0.0","kind":"app","uses":["core"]}')
  const client = await connect()
  const { projects } = await client.call('discover', { manifests: ['libs/core/widget.json', 'apps/shop/widget.json'] })
  assert.equal(projects.length, 2)
  const [core, shop] = projects
  assert.equal(core?.kind, 'lib')
  assert.equal(shop?.kind, 'app')
  assert.deepEqual(shop?.imports, ['core'])
  assert.deepEqual(core?.exports, ['core'])
  assert.equal(core?.tasks?.build?.outputs?.[0], '../../dist/libs/core')
  await client.close()
})

test('a dry run of version/write changes nothing', async () => {
  write('libs/core/widget.json', '{\n  "name": "core",\n  "version": "1.0.0"\n}\n')
  const client = await connect()
  const dry = await client.call('version/write', { project: core, version: '1.1.0', dryRun: true })
  assert.deepEqual(dry.files, ['libs/core/widget.json'])
  assert.match(readFileSync(join(workspace, 'libs', 'core', 'widget.json'), 'utf8'), /"1\.0\.0"/)
  await client.call('version/write', { project: core, version: '1.1.0' })
  assert.match(readFileSync(join(workspace, 'libs', 'core', 'widget.json'), 'utf8'), /"version": "1\.1\.0"/)
  await client.close()
})

test('doctor offers a fix for a missing license', async () => {
  write('libs/core/widget.json', '{"name":"core","version":"1.0.0"}')
  const client = await connect()
  const { findings } = await client.call('doctor', { projects: [core] })
  const finding = findings[0]
  assert.equal(finding?.ok, false)
  assert.ok(finding.remedy)
  assert.equal(finding.fix?.safe, false)
  assert.match(finding.fix.files?.[0]?.content ?? '', /"license": "MIT"/)
  await client.close()
})
