/* eslint-disable unicorn/no-process-exit -- a command line program: it ends the process when smith would */
'use strict'
// A stand-in for `smith serve`, for the integration tests: it speaks the
// JSON-RPC and event protocol of docs/reference/serve.md for a fixed workspace
// of two projects, so the tests do not depend on a toolchain being installed.
// The task `hang` of project `lib` never ends by itself; only `cancel` ends it.
if (process.argv[2] === '--version') {
  process.stdout.write('smith version 9.9.9\n')
  process.exit(0)
}

const fs = require('node:fs')
const path = require('node:path')
const readline = require('node:readline')

// `smith agent install`, `smith mcp install` and `smith config set agent.enabled false`, as far as the
// extension looks at them: the previews, the refusals and the files of docs/manual/agent-install.md.
const AGENT_FILES = ['.claude/agents/smith.md', '.mcp.json', '.vscode/mcp.json']
const [command, subcommand, ...rest] = process.argv.slice(2)
const stop = (message, hint) => {
  process.stderr.write(`✖ ${message}\n  → ${hint}\n`)
  process.exit(1)
}
if (subcommand === 'install' && (command === 'agent' || command === 'mcp')) {
  const wanted = command === 'agent' ? AGENT_FILES : ['.vscode/mcp.json']
  const missing = wanted.filter((file) => !fs.existsSync(file))
  const edited = wanted.find((file) => fs.existsSync(file) && fs.readFileSync(file, 'utf8').includes('EDITED'))
  if (edited !== undefined) {
    process.stdout.write(`● ${edited} differs from what smith writes\n`)
    stop(`${edited} was changed after smith wrote it, so smith left everything as it is`, 'keep your version: nothing to do')
  }
  if (missing.length === 0) {
    process.stdout.write(`✓ The MonoSmith agent is already installed (${wanted.length} files)\n`)
    process.exit(0)
  }
  if (!rest.includes('--yes')) {
    for (const file of missing) process.stdout.write(`● Add ${file}\n--- a/${file}\n+++ b/${file}\n@@ -1,0 +1,1 @@\n+{}\n\n`)
    stop('nothing was written: there is no terminal to ask on', 'run it again with --yes to write the files above')
  }
  for (const file of missing) {
    fs.mkdirSync(path.dirname(file), { recursive: true })
    fs.writeFileSync(file, '{}\n')
    process.stdout.write(`✓ Wrote ${file}\n`)
  }
  process.exit(0)
}
if (command === 'config' && subcommand === 'set' && rest[0] === 'agent.enabled') {
  fs.appendFileSync('smith.toml', `\n[agent]\nenabled = ${rest[1]}\n`)
  process.stdout.write(`✔ agent.enabled = ${rest[1]}\n`)
  process.exit(0)
}
// The Health view asks `smith derive --check`; this workspace has not turned derived files on.
if (command === 'derive') stop('derived files are off in this workspace', 'set enabled = true under [derive] in smith.toml')
const optedOut = () => fs.existsSync('smith.toml') && fs.readFileSync('smith.toml', 'utf8').includes('enabled = false')
const agentFinding = () => ({
  check:   'The MonoSmith agent files are installed and current',
  ok:      AGENT_FILES.every((file) => fs.existsSync(file)),
  warning: true,
  detail:  'missing files',
})

const PROJECTS = [
  { name: 'demo', dir: 'apps/demo', kind: 'app', language: 'node', tasks: ['build', 'test'] },
  { name: 'lib', dir: 'libs/lib', kind: 'lib', language: 'go', tasks: ['hang'] },
]

const rl = readline.createInterface({ input: process.stdin })
const running = new Map()
let seq = 0

const send = (message) => process.stdout.write(`${JSON.stringify({ jsonrpc: '2.0', ...message })}\n`)
const event = (run, body) => send({ method: 'event', params: { run, seq: ++seq, time: new Date().toISOString(), ...body } })

async function run (id, params) {
  const runId = params.run ?? `fake-${Date.now()}`
  const wanted = new Set(params.tasks)
  const tasks = PROJECTS
    .filter((project) => params.projects === undefined || params.projects.includes(project.name))
    .flatMap((project) => project.tasks.filter((task) => wanted.has(task)).map((task) => `${project.name}:${task}`))
  const state = { cancelled: false, wake: undefined }
  running.set(runId, state)
  event(runId, { event: 'run.started', command: `smith run ${params.tasks.join(' ')}`, tasks })
  for (const task of tasks) event(runId, { event: 'task.queued', task })
  let failed = 0
  for (const task of tasks) {
    event(runId, { event: 'task.started', task })
    event(runId, { event: 'task.output', task, stream: 'combined', line: `running ${task}` })
    if (task.endsWith(':hang')) await new Promise((resolve) => { state.wake = resolve })
    const status = state.cancelled ? 'failed' : 'passed'
    if (status === 'failed') failed++
    event(runId, { event: 'task.finished', task, status, durationMs: 12, ...(state.cancelled && { reason: 'cancelled' }) })
    if (state.cancelled) break
  }
  event(runId, { event: 'run.finished', ok: failed === 0, summary: { passed: tasks.length - failed, failed, cached: 0, skipped: 0, durationMs: 12 } })
  running.delete(runId)
  send({ id, result: { run: runId, ok: failed === 0 } })
}

rl.on('line', (line) => {
  const request = JSON.parse(line)
  const { id, method, params = {} } = request
  switch (method) {
    case 'initialize': { send({ id, result: { name: 'smith', version: '9.9.9', protocolVersion: 1, eventVersion: 1, methods: [] } }); break }
    case 'projects': { send({ id, result: PROJECTS }); break }
    case 'affected': { send({ id, result: { projects: ['demo'] } }); break }
    case 'doctor': { send({ id, result: { findings: [{ check: 'the cache directory is ignored by git', ok: true }, ...(optedOut() ? [] : [agentFinding()])] } }); break }
    case 'updates': { send({ id, result: { updates: [], tools: [] } }); break }
    case 'release.preview': { send({ id, result: { dryRun: true, results: [], unchanged: ['demo'] } }); break }
    case 'run': { void run(id, params); break }
    case 'cancel': {
      const state = running.get(params.run)
      if (state !== undefined) {
        state.cancelled = true
        state.wake?.()
      }
      send({ id, result: { cancelled: state !== undefined } })
      break
    }
    case 'shutdown': {
      send({ id, result: {} })
      process.exit(0)
      break
    }
    default: { if (id !== undefined) send({ id, error: { code: -32601, message: `method not found: ${method}` } }) }
  }
})
rl.on('close', () => process.exit(0))
