import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'

import { chooseVersion, compareVersions, pinIn } from './choose-version.policy.mjs'

test('the pin is [workspace] smith, not a smith key of another table', () => {
  assert.equal(pinIn('[workspace]\nname = "acme"\nsmith = "2.50.0" # pinned\n'), '2.50.0')
  assert.equal(pinIn('[ci]\nsmith = "2.50.0"\n[workspace]\nname = "acme"\n'), null)
  assert.equal(pinIn('[workspace]\nsmith = "latest"\n'), null)
  assert.equal(pinIn('[workspace]\nsmith_version = "2.50.0"\n'), null)
})

test('SMITH_VERSION wins over the pin of the nearest smith.toml above the directory', () => {
  const root = mkdtempSync(join(tmpdir(), 'launcher-pin-'))
  writeFileSync(join(root, 'smith.toml'), '[workspace]\nname = "acme"\nsmith = "2.49.0"\n')
  const deep = join(root, 'apps', 'api')
  mkdirSync(deep, { recursive: true })
  assert.equal(chooseVersion({}, deep), '2.49.0')
  assert.equal(chooseVersion({ SMITH_VERSION: '2.50.1' }, deep), '2.50.1')
  assert.equal(chooseVersion({ SMITH_VERSION: 'latest' }, deep), null)
  assert.throws(() => chooseVersion({ SMITH_VERSION: 'two' }, deep), /SMITH_VERSION is "two"/)
})

test('versions compare by number, and a prerelease comes before its release', () => {
  assert.ok(compareVersions('2.10.0', '2.9.9') > 0)
  assert.ok(compareVersions('3.0.0-beta.1', '3.0.0') < 0)
  assert.equal(compareVersions('2.50.0', '2.50.0'), 0)
})
