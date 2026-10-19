#!/usr/bin/env node
// A minimal smith language plugin in Node.js, no dependencies.
//
// It teaches smith one made-up project type: every directory that holds a
// widget.json is a project, and each gets a "hello" task.
//
// The protocol is JSON-RPC 2.0 on stdin and stdout, each message framed with a
// "Content-Length: <bytes>" header, a blank line and the JSON body (the framing
// of the Language Server Protocol). Anything for a human goes to stderr.
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

const PROTOCOL_VERSION = 1
let workspaceRoot = ''

const handlers = {
  initialize (params) {
    workspaceRoot = params.workspaceRoot
    if (!params.protocolVersions.includes(PROTOCOL_VERSION)) {
      throw new Error(`smith speaks ${params.protocolVersions}, this plugin speaks ${PROTOCOL_VERSION}`)
    }
    return {
      protocolVersion: PROTOCOL_VERSION,
      plugin: { name: 'widget', version: '0.1.0' },
      kind: 'language',
      roles: ['language'], // not also scaffold: this plugin does not generate
      languages: ['widget'],
      manifests: ['widget.json'],
      // Only what the plugin implements; smith never calls anything else.
      capabilities: { discover: true }
    }
  },

  discover (params) {
    const projects = []
    for (const manifest of params.manifests) { // workspace-relative, forward slashes
      const data = JSON.parse(readFileSync(join(workspaceRoot, ...manifest.split('/')), 'utf8'))
      if (!data.name) continue // not ours
      projects.push({
        name: data.name,
        root: dirname(manifest).replaceAll('\\', '/'),
        language: 'widget',
        kind: data.kind || 'lib',
        version: data.version || '',
        exports: [data.name],
        tasks: {
          hello: {
            command: ['git', '--version'],
            inputs: ['widget.json'],
            cache: true,
            description: 'say hello, as a widget does'
          }
        }
      })
    }
    return { projects }
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

// Cut the byte stream into framed messages.
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
