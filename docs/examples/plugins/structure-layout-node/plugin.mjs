#!/usr/bin/env node
// A smith structure plugin in Node.js, no dependencies: it adds a
// layout-check task to widget projects and makes their lint wait for it.
//
// Started by smith (SMITH_PLUGIN=1) it speaks the plugin protocol on stdin and
// stdout. Started with `layout-check`, which is the command of the task it
// contributes, it checks the folder it runs in. Anything meant for a human
// goes to stderr: stdout carries only protocol frames.
import { existsSync, readdirSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

const PROTOCOL_VERSION = 1
const JUNK_DRAWERS = ['util', 'utils', 'helper', 'helpers', 'common', 'shared', 'misc']
const SKIPPED = ['dist', 'node_modules', '.git']
let workspaceRoot = ''

// findJunkDrawers lists the folders below dir named for nothing, as
// slash-separated paths relative to dir.
export function findJunkDrawers (dir, prefix = '') {
  const found = []
  for (const entry of readdirSync(join(dir, prefix), { withFileTypes: true })) {
    if (!entry.isDirectory() || SKIPPED.includes(entry.name)) continue
    const rel = prefix ? `${prefix}/${entry.name}` : entry.name
    if (JUNK_DRAWERS.includes(entry.name.toLowerCase())) found.push(rel)
    found.push(...findJunkDrawers(dir, rel))
  }
  return found
}

const handlers = {
  initialize (params) {
    workspaceRoot = params.workspaceRoot
    if (!params.protocolVersions.includes(PROTOCOL_VERSION)) {
      throw new Error(`smith speaks ${params.protocolVersions}, this plugin speaks ${PROTOCOL_VERSION}`)
    }
    return {
      protocolVersion: PROTOCOL_VERSION,
      plugin: { name: 'widget-layout', version: '0.1.0' },
      kind: 'convention',
      roles: ['structure'],
      languages: ['widget'], // a convention that names none applies to no project
      // Only what the plugin implements; smith never calls anything else.
      capabilities: { contribute: true, doctor: true }
    }
  },

  contribute (params) {
    const command = ['node', fileURLToPath(import.meta.url), 'layout-check']
    const projects = {}
    for (const project of params.projects) {
      projects[project.name] = {
        tasks: {
          'layout-check': {
            command,
            cache: true,
            description: 'fail on folders named util, helpers, common, shared or misc'
          }
        },
        // Ignored for a project that has no lint once every plugin has contributed.
        dependsOn: { lint: ['layout-check'] }
      }
    }
    return { projects }
  },

  doctor (params) {
    return {
      findings: params.projects.map((project) => {
        const ok = existsSync(join(workspaceRoot, ...project.root.split('/'), 'README.md'))
        const finding = { check: `${project.name} has a README.md`, ok }
        if (!ok) {
          finding.detail = `${project.root}/README.md does not exist`
          finding.remedy = `write ${project.root}/README.md: what the project is for`
          finding.warning = true // reported, but does not fail smith doctor
        }
        return finding
      })
    }
  },

  shutdown () { return {} }
}

function send (message) {
  const body = Buffer.from(JSON.stringify(message), 'utf8')
  process.stdout.write(Buffer.concat([Buffer.from(`Content-Length: ${body.length}\r\n\r\n`, 'ascii'), body]))
}

function handle (request) {
  if (request.id === undefined) return // a notification: no reply
  const response = { jsonrpc: '2.0', id: request.id }
  const handler = handlers[request.method]
  if (!handler) {
    response.error = { code: -32601, message: `method not supported: ${request.method}` }
  } else {
    try {
      response.result = handler(request.params ?? {})
    } catch (e) { // -32000: a failure the user must act on
      response.error = { code: -32000, message: e.message }
    }
  }
  send(response)
  if (request.method === 'shutdown') process.exit(0)
}

function serve () {
  // Cut the byte stream into framed messages. Count bytes, not characters.
  let buffer = Buffer.alloc(0)
  process.stdin.on('data', (chunk) => {
    buffer = Buffer.concat([buffer, chunk])
    for (;;) {
      const end = buffer.indexOf('\r\n\r\n')
      if (end < 0) return
      const match = /content-length:\s*(\d+)/i.exec(buffer.subarray(0, end).toString('ascii'))
      if (!match) throw new Error('message without Content-Length')
      const length = Number(match[1])
      if (buffer.length < end + 4 + length) return
      const body = buffer.subarray(end + 4, end + 4 + length).toString('utf8')
      buffer = buffer.subarray(end + 4 + length)
      handle(JSON.parse(body))
    }
  })
}

if (process.env.SMITH_PLUGIN === '1') {
  serve()
} else if (process.argv[2] === 'layout-check') {
  const found = findJunkDrawers('.')
  if (found.length > 0) {
    console.error(`folders named for nothing: ${found.join(', ')} (name them for what they do)`)
    process.exit(1)
  }
  console.log('no junk-drawer folders')
} else if (process.argv[1] === fileURLToPath(import.meta.url)) {
  console.error('usage: plugin.mjs layout-check (smith runs this program as a plugin)')
  process.exit(1)
}
