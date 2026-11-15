#!/usr/bin/env node
// The smith command npm installs: a launcher, not smith itself (ADR 0101). It runs the smith the
// workspace pins in smith.toml ([workspace] smith), SMITH_VERSION, or the newest installed, and
// downloads and checks that version the first time.
import { readFileSync } from 'node:fs'

import { resolveProgram, runSmith } from './smith-run/index.mjs'

const { version } = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'))
const say = line => process.stderr.write(`smith launcher: ${line}\n`)

try {
  const program = await resolveProgram({ env: process.env, cwd: process.cwd(), platform: process.platform, arch: process.arch, say })
  process.exitCode = runSmith(program, process.argv.slice(2), process.env, `npm@${version}`)
} catch (error) {
  say(error.message)
  process.exitCode = 1
}
