import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { mkdirSync, mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { findJunkDrawers } from './plugin.mjs'

const plugin = fileURLToPath(new URL('./plugin.mjs', import.meta.url))

// talk starts the plugin the way smith does and plays a conversation: each
// request is a frame on stdin, each answer a frame on stdout.
function talk (workspaceRoot, requests) {
  return new Promise((resolve, reject) => {
    const child = spawn(process.execPath, [plugin], { env: { ...process.env, SMITH_PLUGIN: '1' } })
    const chunks = []
    child.stdout.on('data', (c) => chunks.push(c))
    child.on('error', reject)
    child.on('close', () => {
      const out = Buffer.concat(chunks)
      const answers = []
      for (let i = 0; i < out.length;) {
        const end = out.indexOf('\r\n\r\n', i)
        const length = Number(/content-length:\s*(\d+)/i.exec(out.subarray(i, end).toString('ascii'))[1])
        answers.push(JSON.parse(out.subarray(end + 4, end + 4 + length).toString('utf8')))
        i = end + 4 + length
      }
      resolve(answers)
    })
    const all = [
      { method: 'initialize', params: { protocolVersions: [1], host: { name: 'test', version: '0' }, workspaceRoot } },
      ...requests,
      { method: 'shutdown', params: {} }
    ]
    all.forEach((r, i) => {
      const body = Buffer.from(JSON.stringify({ jsonrpc: '2.0', id: i + 1, ...r }))
      child.stdin.write(Buffer.concat([Buffer.from(`Content-Length: ${body.length}\r\n\r\n`), body]))
    })
  })
}

test('the handshake advertises exactly what is implemented', async () => {
  const [init] = await talk(tmpdir(), [])
  assert.equal(init.result.protocolVersion, 1)
  assert.deepEqual(init.result.roles, ['structure'])
  assert.deepEqual(init.result.capabilities, { contribute: true, doctor: true })
})

test('contribute adds layout-check and makes lint wait for it', async () => {
  const [, contribute] = await talk(tmpdir(), [{ method: 'contribute', params: { projects: [{ name: 'core', root: 'libs/core' }] } }])
  const core = contribute.result.projects.core
  assert.deepEqual(core.dependsOn, { lint: ['layout-check'] })
  assert.equal(core.tasks['layout-check'].command.at(-1), 'layout-check')
})

test('an unknown method is -32601', async () => {
  const [, nope] = await talk(tmpdir(), [{ method: 'publish', params: {} }])
  assert.equal(nope.error.code, -32601)
})

test('doctor warns about a missing README', async () => {
  const ws = mkdtempSync(join(tmpdir(), 'layout-'))
  try {
    mkdirSync(join(ws, 'libs', 'core'), { recursive: true })
    const [, doctor] = await talk(ws, [{ method: 'doctor', params: { projects: [{ name: 'core', root: 'libs/core' }] } }])
    const finding = doctor.result.findings[0]
    assert.equal(finding.ok, false)
    assert.equal(finding.warning, true)
    assert.ok(finding.remedy)
  } finally {
    rmSync(ws, { recursive: true, force: true })
  }
})

test('findJunkDrawers skips build output', () => {
  const dir = mkdtempSync(join(tmpdir(), 'layout-'))
  try {
    for (const d of ['src/orders', 'src/Utils', 'dist/helpers', 'common']) mkdirSync(join(dir, ...d.split('/')), { recursive: true })
    assert.deepEqual(findJunkDrawers(dir).sort(), ['common', 'src/Utils'])
  } finally {
    rmSync(dir, { recursive: true, force: true })
  }
})
