'use strict'
// Tests of the launcher's choice logic, with fake npm commands and no clock.
// Run: node --test apps/smith/npmbinaries
const test = require('node:test')
const assert = require('node:assert')
const fs = require('node:fs')
const path = require('node:path')
const Module = require('node:module')

function load() {
  const file = path.join(__dirname, 'shim_template.js')
  const source = fs.readFileSync(file, 'utf8')
    .replace(/^#!.*\n/, '')
    .replaceAll('__BINARY__', 'smith')
    .replace('__PLATFORMS__', JSON.stringify({ 'linux-x64': '@acme/cli-linux-x64' }))
  const m = new Module(file)
  m.filename = file
  m.paths = Module._nodeModulePaths(__dirname)
  m._compile(source, file)
  return m.exports
}

const { resolveBinary } = load()

// deps builds fake dependencies. registry says what `npm view` answers.
function deps(overrides = {}) {
  const calls = []
  const lines = []
  const questions = []
  const answers = [...(overrides.answers ?? [])]
  let installed = false
  const d = {
    key: 'linux-x64', version: '1.2.3', env: {}, isTTY: true,
    log: (l) => lines.push(l),
    resolve: () => (installed ? '/global/@acme/cli-linux-x64/bin/smith' : null),
    npm: (args) => {
      calls.push(args.join(' '))
      if (args[0] === 'install') {
        installed = overrides.installWorks !== false
        return { status: installed ? 0 : 1 }
      }
      return (overrides.view ?? (() => ({ status: 0, stdout: '1.2.3\n', stderr: '' })))(args, calls)
    },
    ask: async (q) => { questions.push(q); return answers.shift() ?? '' },
    sleep: async () => { d.slept = (d.slept ?? 0) + 1 },
    ...overrides.deps,
  }
  return { d, calls, lines, questions }
}

test('the binary is found when the platform package is installed', async () => {
  const { d, calls } = deps({ deps: { resolve: () => '/x/bin/smith' } })
  assert.strictEqual(await resolveBinary(d), '/x/bin/smith')
  assert.deepStrictEqual(calls, [])
})

test('a platform with no build is said so', async () => {
  const { d, lines } = deps({ deps: { key: 'sunos-x64' } })
  assert.strictEqual(await resolveBinary(d), null)
  assert.match(lines.join('\n'), /no build for sunos-x64/)
})

test('(a) the package is on the registry: it offers the command and installs only on yes', async () => {
  const yes = deps({ answers: [''] })
  assert.strictEqual(await resolveBinary(yes.d), '/global/@acme/cli-linux-x64/bin/smith')
  assert.deepStrictEqual(yes.calls, ['view @acme/cli-linux-x64@1.2.3 version', 'install -g @acme/cli-linux-x64@1.2.3'])
  assert.match(yes.lines.join('\n'), /to fix it, run: npm install -g @acme\/cli-linux-x64@1\.2\.3/)
  assert.match(yes.questions[0], /Install it now\? \[Y\/n\]/)

  const no = deps({ answers: ['n'] })
  assert.strictEqual(await resolveBinary(no.d), null)
  assert.ok(!no.calls.some((c) => c.startsWith('install')), 'no install without a yes')
})

test('(a) without a terminal nothing is installed, and the command is printed', async () => {
  const { d, calls, lines, questions } = deps({ deps: { isTTY: false } })
  assert.strictEqual(await resolveBinary(d), null)
  assert.deepStrictEqual(calls, ['view @acme/cli-linux-x64@1.2.3 version'])
  assert.deepStrictEqual(questions, [])
  assert.match(lines.join('\n'), /npm install -g @acme\/cli-linux-x64@1\.2\.3/)
})

test('(a) SMITH_AUTO_FIX=1 accepts the install without a terminal', async () => {
  const { d, calls } = deps({ deps: { isTTY: false, env: { SMITH_AUTO_FIX: '1' } } })
  assert.strictEqual(await resolveBinary(d), '/global/@acme/cli-linux-x64/bin/smith')
  assert.ok(calls.includes('install -g @acme/cli-linux-x64@1.2.3'))
})

test('(a) a failed install says so and returns null', async () => {
  const { d, lines } = deps({ answers: ['y'], installWorks: false })
  assert.strictEqual(await resolveBinary(d), null)
  assert.match(lines.join('\n'), /install failed/)
})

const absent = { view: () => ({ status: 0, stdout: '', stderr: '' }) }

test('(b) not on the registry yet: it says try again, and waits only when asked', async () => {
  const no = deps({ ...absent, answers: [''] })
  assert.strictEqual(await resolveBinary(no.d), null)
  assert.match(no.lines.join('\n'), /not on the npm registry yet.*Try again in a few minutes/)
  assert.match(no.questions[0], /\[y\/N\]/, 'the default is marked, and it is no')
  assert.strictEqual(no.d.slept, undefined)
})

test('(b) a 404 from npm is the same situation', async () => {
  const { d, lines } = deps({ view: () => ({ status: 1, stdout: '', stderr: 'npm error code E404' }), deps: { isTTY: false } })
  assert.strictEqual(await resolveBinary(d), null)
  assert.match(lines.join('\n'), /not on the npm registry yet/)
})

test('(b) waiting retries every 30 s and offers the install when the package appears', async () => {
  let views = 0
  const { d, calls, questions } = deps({
    answers: ['y', 'y'],
    view: () => { views++; return views < 4 ? { status: 0, stdout: '' } : { status: 0, stdout: '1.2.3\n' } },
  })
  assert.strictEqual(await resolveBinary(d), '/global/@acme/cli-linux-x64/bin/smith')
  assert.strictEqual(d.slept, 3)
  assert.strictEqual(calls.filter((c) => c.startsWith('view')).length, 4)
  assert.match(questions[0], /Wait and retry every 30 seconds/)
  assert.match(questions[1], /Install it now/)
})

test('(b) waiting stops after 10 minutes', async () => {
  const { d, lines } = deps({ ...absent, answers: ['y'] })
  assert.strictEqual(await resolveBinary(d), null)
  assert.strictEqual(d.slept, 20)
  assert.match(lines.join('\n'), /still not on the registry after 10 minutes/)
})

test('(c) npm cannot be asked: it says so and installs nothing', async () => {
  for (const view of [
    () => ({ error: new Error('ENOENT') }),
    () => ({ status: null, signal: 'SIGTERM' }),
    () => ({ status: 1, stderr: 'npm error network request failed' }),
  ]) {
    const { d, calls, lines, questions } = deps({ view })
    assert.strictEqual(await resolveBinary(d), null)
    assert.match(lines.join('\n'), /npm could not be asked/)
    assert.deepStrictEqual(calls.length, 1)
    assert.deepStrictEqual(questions, [])
  }
})
