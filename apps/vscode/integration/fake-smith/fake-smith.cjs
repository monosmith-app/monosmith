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

const readline = require('node:readline')

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
    case 'doctor': { send({ id, result: { findings: [{ check: 'the cache directory is ignored by git', ok: true }] } }); break }
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
