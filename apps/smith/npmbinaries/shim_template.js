#!/usr/bin/env node
'use strict'
// Runs the __BINARY__ binary npm installed for this platform, from one of the
// optional dependencies; nothing is downloaded after install. When that package
// is missing, it says why and offers a fix, and changes nothing without a yes.
const { spawn, spawnSync } = require('node:child_process')
const fs = require('node:fs')

const platforms = __PLATFORMS__
const NAME = '__BINARY__'
const PROBE_TIMEOUT_MS = 5000
const RETRY_EVERY_MS = 30000
const RETRY_FOR_MS = 10 * 60 * 1000

// probe asks npm, with a short limit, whether pkg@version is on the registry.
// It returns 'exists', 'absent' or 'unknown' (npm could not be asked).
function probe(d, pkg) {
  const r = d.npm(['view', pkg + '@' + d.version, 'version'])
  if (r.error || r.signal || r.status === null) return 'unknown'
  if (r.status === 0) return String(r.stdout).trim() === d.version ? 'exists' : 'absent'
  return /E404|404 Not Found|is not in this registry/i.test(String(r.stderr)) ? 'absent' : 'unknown'
}

// ask puts a question and returns true for yes. Enter takes the default.
async function ask(d, question, defaultYes) {
  const answer = String(await d.ask(question + (defaultYes ? ' [Y/n] ' : ' [y/N] '))).trim().toLowerCase()
  if (answer === '') return defaultYes
  return answer === 'y' || answer === 'yes'
}

// install runs the exact command it printed, and says if it worked.
function install(d, pkg) {
  const r = d.npm(['install', '-g', pkg + '@' + d.version], { stdio: 'inherit', timeout: 0 })
  return !r.error && r.status === 0
}

// offerInstall is situation (a): the package is on the registry now.
async function offerInstall(d, pkg) {
  const command = 'npm install -g ' + pkg + '@' + d.version
  d.log(NAME + ': ' + pkg + '@' + d.version + ' is on the npm registry, so your install missed it (an interrupted or partial install).')
  d.log(NAME + ': to fix it, run: ' + command)
  let yes = false
  if (d.env.SMITH_AUTO_FIX === '1') {
    d.log(NAME + ': SMITH_AUTO_FIX=1 is set, so it runs now.')
    yes = true
  } else if (d.isTTY) {
    yes = await ask(d, 'Install it now?', true)
  }
  if (!yes) return null
  if (!install(d, pkg)) {
    d.log(NAME + ': the install failed; run the command above yourself to see why.')
    return null
  }
  return d.resolve(pkg)
}

// waitAndRetry is situation (b) when the person agrees to wait.
async function waitAndRetry(d, pkg) {
  for (let waited = 0; waited < RETRY_FOR_MS; waited += RETRY_EVERY_MS) {
    d.log(NAME + ': waiting for ' + pkg + '@' + d.version + ' to reach the registry (' + Math.round(waited / 1000) + ' s of ' + RETRY_FOR_MS / 1000 + ' s); press Ctrl+C to stop.')
    await d.sleep(RETRY_EVERY_MS)
    const state = probe(d, pkg)
    if (state === 'exists') return offerInstall(d, pkg)
    if (state === 'unknown') {
      d.log(NAME + ': npm could not be asked any more, so it stops waiting.')
      return null
    }
  }
  d.log(NAME + ': ' + pkg + '@' + d.version + ' is still not on the registry after ' + RETRY_FOR_MS / 60000 + ' minutes. Try again later.')
  return null
}

// resolveBinary returns the path of the binary of this platform. When the
// platform package is missing it tells which of three situations this is
// (on the registry now, not there yet, or npm cannot be asked) and returns
// null unless a fix was agreed to and worked.
async function resolveBinary(d) {
  const pkg = platforms[d.key]
  if (!pkg) {
    d.log(NAME + ' has no build for ' + d.key)
    return null
  }
  const found = d.resolve(pkg)
  if (found) return found
  d.log(NAME + ': ' + pkg + ' is not installed. npm skips an optional package without an error when it cannot get it.')
  switch (probe(d, pkg)) {
    case 'exists':
      return offerInstall(d, pkg)
    case 'absent':
      d.log(NAME + ': ' + pkg + '@' + d.version + ' is not on the npm registry yet. A new release can take 5 to 10 minutes to spread. Try again in a few minutes.')
      if (d.isTTY && await ask(d, 'Wait and retry every 30 seconds, for up to 10 minutes?', false)) return waitAndRetry(d, pkg)
      return null
    default:
      d.log(NAME + ': npm could not be asked (it is not in PATH, it did not answer in ' + PROBE_TIMEOUT_MS / 1000 + ' s, or there is no network).')
      d.log(NAME + ': when you are online, run: npm install -g ' + pkg + '@' + d.version + ' (or reinstall without --omit=optional or --no-optional)')
      return null
  }
}

// realDeps is what the launcher uses; tests pass their own.
function realDeps() {
  const file = process.platform === 'win32' ? NAME + '.exe' : NAME
  const isTTY = Boolean(process.stdin.isTTY && process.stderr.isTTY)
  return {
    key: process.platform + '-' + process.arch,
    version: require('../package.json').version,
    env: process.env,
    isTTY,
    log: (line) => console.error(line),
    resolve: (pkg) => {
      try { return require.resolve(pkg + '/bin/' + file) } catch { return null }
    },
    // npm is a .cmd file on Windows, which needs a shell to start.
    npm: (args, options) => spawnSync('npm', args, { encoding: 'utf8', timeout: PROBE_TIMEOUT_MS, shell: process.platform === 'win32', ...options }),
    ask: (question) => new Promise((resolve) => {
      const rl = require('node:readline').createInterface({ input: process.stdin, output: process.stderr })
      rl.question(question, (answer) => { rl.close(); resolve(answer) })
    }),
    sleep: (ms) => new Promise((resolve) => setTimeout(resolve, ms)),
  }
}

function launch(binary) {
  if (process.platform !== 'win32') {
    try {
      fs.accessSync(binary, fs.constants.X_OK)
    } catch {
      try { fs.chmodSync(binary, 0o755) } catch { /* a read-only install keeps its mode */ }
    }
  }
  // A stop request must reach the binary: forward it, never orphan the child.
  const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit' })
  for (const sig of ['SIGINT', 'SIGTERM', 'SIGHUP']) {
    process.on(sig, () => { child.kill(sig) })
  }
  child.on('error', (error) => {
    console.error(NAME + ': ' + error.message)
    process.exit(1)
  })
  child.on('exit', (code, signal) => {
    if (signal) {
      process.removeAllListeners(signal)
      process.kill(process.pid, signal)
      return
    }
    process.exit(code ?? 1)
  })
}

module.exports = { resolveBinary, probe }

if (require.main === module) {
  resolveBinary(realDeps()).then((binary) => {
    if (binary) launch(binary)
    else process.exit(1)
  })
}
