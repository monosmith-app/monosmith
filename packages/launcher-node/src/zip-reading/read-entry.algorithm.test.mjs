import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { test } from 'node:test'

import { readEntry } from './read-entry.algorithm.mjs'

const zip = readFileSync(new URL('../smith-run/fixtures/monosmith_1.2.3_linux_amd64.zip', import.meta.url))

test('one entry of a deflated zip, as Go writes release archives', () => {
  assert.equal(readEntry(zip, name => name === 'smith').toString('utf8'), '#!/bin/sh\necho smith 1.2.3\n')
  assert.equal(readEntry(zip, name => name === 'missing'), null)
  assert.throws(() => readEntry(Buffer.from('not a zip'), () => true), /not a zip/)
})
