'use strict'
// The entry VS Code loads in the extension host (--extensionTestsPath): it
// runs every spec file of the bundled suite that MONOSMITH_SUITE_DIR names.
const { readdirSync } = require('node:fs')
const { join } = require('node:path')
const Mocha = require('mocha')

async function run () {
  const dir = process.env.MONOSMITH_SUITE_DIR
  if (dir === undefined) throw new Error('MONOSMITH_SUITE_DIR is not set')
  const mocha = new Mocha({ ui: 'bdd', color: true, timeout: 60_000 })
  for (const file of readdirSync(dir)) {
    if (file.endsWith('.integration.spec.js')) mocha.addFile(join(dir, file))
  }

  await new Promise((resolve, reject) => {
    mocha.run((failures) => { failures > 0 ? reject(new Error(`${failures} integration tests failed`)) : resolve() })
  })
}

module.exports = { run }
