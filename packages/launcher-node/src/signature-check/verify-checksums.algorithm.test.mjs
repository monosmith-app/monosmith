import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { test } from 'node:test'

import { checkArchive, listedDigest } from './verify-checksums.algorithm.mjs'

// A release signed by smith release's own minisign code (aead.dev/minisign), with a test key.
const fixture = name => readFileSync(new URL(`../smith-run/fixtures/${name}`, import.meta.url))
const publicKey = fixture('test-key.pub').toString('utf8').split('\n', 2)[1]
const download = () => ({
  archive:   fixture('monosmith_1.2.3_linux_amd64.zip'),
  name:      'monosmith_1.2.3_linux_amd64.zip',
  checksums: fixture('checksums.txt'),
  signature: fixture('checksums.txt.minisig').toString('utf8'),
})

test('an archive listed in a correctly signed checksums.txt is trusted', () => {
  assert.equal(checkArchive(download(), publicKey), null)
})

test('a changed list, another key or a changed archive is not trusted', () => {
  const changed = Buffer.concat([download().checksums, Buffer.from('x')])
  assert.match(checkArchive({ ...download(), checksums: changed }, publicKey), /signature/)
  assert.match(checkArchive(download()), /signature/)
  assert.match(checkArchive({ ...download(), archive: Buffer.from('other') }, publicKey), /does not match its checksum/)
  assert.match(checkArchive({ ...download(), name: 'monosmith_1.2.3_darwin_arm64.zip' }, publicKey), /does not list/)
})

test('a digest is found by the file name of its line', () => {
  assert.equal(listedDigest('ab12  one.zip\ncd34  two.zip\n', 'two.zip'), 'cd34')
  assert.equal(listedDigest('ab12  one.zip\n', 'three.zip'), null)
})
