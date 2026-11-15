import assert from 'node:assert/strict'
import { existsSync, mkdtempSync, readFileSync } from 'node:fs'
import { createServer } from 'node:http'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { after, before, test } from 'node:test'

import { resolveProgram } from './run-smith.use-case.mjs'

// A mirror of one release (SMITH_DOWNLOAD_URL), serving the files smith release attaches.
const fixture = name => readFileSync(new URL(`fixtures/${name}`, import.meta.url))
const publicKey = fixture('test-key.pub').toString('utf8').split('\n', 2)[1]
const requests = []
let server
let base

before(async () => {
  server = createServer((req, res) => {
    requests.push(req.url)
    const files = { '/latest.txt': Buffer.from('1.2.3\n') }
    for (const name of ['monosmith_1.2.3_linux_amd64.zip', 'checksums.txt', 'checksums.txt.minisig']) files[`/smith%401.2.3/${name}`] = fixture(name)
    const body = files[req.url]
    res.writeHead(body ? 200 : 404)
    res.end(body)
  })
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve))
  base = `http://127.0.0.1:${server.address().port}`
})

after(() => server.close())

const host = (env, cwd) => ({ env, cwd, platform: 'linux', arch: 'x64', say: () => {}, publicKey })

test('the newest release is downloaded once, checked and kept in the versions folder', async () => {
  const home = mkdtempSync(join(tmpdir(), 'launcher-home-'))
  const env = { SMITH_DOWNLOAD_URL: base, SMITH_HOME: home }
  const program = await resolveProgram(host(env, home))
  assert.equal(program, join(home, 'versions', '1.2.3', 'smith'))
  assert.equal(readFileSync(program, 'utf8'), '#!/bin/sh\necho smith 1.2.3\n')
  requests.length = 0
  assert.equal(await resolveProgram(host(env, home)), program)
  assert.deepEqual(requests, [], 'an installed smith runs without the network')
})

test('a version that fails the checks is not installed', async () => {
  const home = mkdtempSync(join(tmpdir(), 'launcher-home-'))
  const env = { SMITH_DOWNLOAD_URL: base, SMITH_HOME: home, SMITH_VERSION: '1.2.3' }
  await assert.rejects(resolveProgram({ ...host(env, home), publicKey: undefined }), /was not installed: the signature/)
  assert.equal(existsSync(join(home, 'versions', '1.2.3')), false)
})

test('a version the mirror lacks says which address answered', async () => {
  const home = mkdtempSync(join(tmpdir(), 'launcher-home-'))
  const env = { SMITH_DOWNLOAD_URL: base, SMITH_HOME: home, SMITH_VERSION: '9.9.9' }
  await assert.rejects(resolveProgram(host(env, home)), /smith%409\.9\.9.* answered 404/)
})
