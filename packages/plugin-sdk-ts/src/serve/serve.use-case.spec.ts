import { PassThrough } from 'node:stream'
import { describe, expect, it } from 'vitest'
import { FrameReader, frameMessage } from '../framing/index.ts'
import type { Plugin, SmithPlugin } from './plugin.contract.ts'
import { serveStreams } from './serve.use-case.ts'

const describeWidget: Plugin['describe'] = () => ({ plugin: { name: 'widget', version: '0.1.0' }, kind: 'language', manifests: ['widget.json'] })
const initialize = { jsonrpc: '2.0', id: 1, method: 'initialize', params: { protocolVersions: [1], host: { name: 'test', version: '0' }, workspaceRoot: '/ws' } }

interface Conversation {
  /** The raw bodies the plugin wrote. */
  bodies:   string[]
  warnings: string[]
}

/** Serves the plugin over pipes, writes the requests (objects, or raw strings framed as they are) and collects what it writes until it ends. */
async function talk (plugin: SmithPlugin, requests: unknown[]): Promise<Conversation> {
  const input = new PassThrough()
  const output = new PassThrough()
  const warnings: string[] = []
  const chunks: Buffer[] = []
  output.on('data', (chunk: Buffer) => {
    chunks.push(chunk)
  })
  for (const request of requests) input.write(frameMessage(typeof request === 'string' ? request : JSON.stringify(request)))
  input.end()
  await serveStreams(plugin, {
    input,
    output,
    log: (message) => {
      warnings.push(message)
    },
  })
  output.end()

  return { bodies: new FrameReader().push(Buffer.concat(chunks)), warnings }
}

const answers = (conversation: Conversation): Array<Record<string, any>> => conversation.bodies.map(body => JSON.parse(body) as Record<string, any>)

describe('the handshake', () => {
  it('answers initialize with the version it chose and the capabilities derived from its methods', async () => {
    const plugin: SmithPlugin = {
      describe:     describeWidget,
      discover:     () => ({ projects: [] }),
      writeVersion: () => ({ files: [] }),
      install:      () => ({}),
      uninstall:    () => ({}),
    }
    const { bodies } = await talk(plugin, [initialize])
    expect(bodies).toEqual(['{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget","version":"0.1.0"},"kind":"language","capabilities":{"discover":true,"version":true,"dependencies":true},"manifests":["widget.json"]}}'])
  })

  it('writes no capabilities a plugin without methods has, and ignores what describe says about them', async () => {
    const meddling = { plugin: { name: 'widget', version: '0.1.0' }, kind: 'language', protocolVersion: 9, capabilities: { publish: true } }
    const plugin: SmithPlugin = { describe: () => meddling as never }
    const [response] = answers(await talk(plugin, [initialize]))
    expect(response.result.capabilities).toEqual({})
    expect(response.result.protocolVersion).toBe(1)
  })

  it('refuses a host that does not speak version 1, with the Go SDK message and code', async () => {
    const [response] = answers(await talk({ describe: describeWidget }, [{ ...initialize, params: { ...initialize.params, protocolVersions: [2, 3] } }]))
    expect(response.error).toEqual({ code: -32000, message: 'no common protocol version: host speaks [2 3], plugin speaks 1' })
  })

  it('refuses every request before initialize, shutdown included, and goes on', async () => {
    const result = answers(await talk({ describe: describeWidget }, [{ jsonrpc: '2.0', id: 1, method: 'shutdown', params: {} }, { jsonrpc: '2.0', id: 2, method: 'discover', params: {} }, initialize]))
    expect(result.map(r => r.error?.code ?? 'ok')).toEqual([-32600, -32600, 'ok'])
    expect(result[0].error.message).toBe('initialize must come first')
  })
})

describe('dispatch', () => {
  const plugin: SmithPlugin = {
    describe: describeWidget,
    discover: (params) => ({ projects: params.manifests.map(manifest => ({ name: manifest, root: manifest, language: 'widget', kind: 'lib' })) }),
    doctor:   () => { throw new Error('the toolchain is missing: install node 24') },
    generate: () => ({ files: 'x' as never }),
    publish:  async () => {
      await new Promise(resolve => setTimeout(resolve, 5))

      return { published: true }
    },
  }
  /** The answer to one request sent after the handshake. */
  const call = async (method: string, params: unknown): Promise<Record<string, any>> => {
    const conversation = await talk(plugin, [initialize, { jsonrpc: '2.0', id: 2, method, params }])

    return answers(conversation)[1]
  }

  it('calls the method with the params shaped as the types promise, awaiting an async one', async () => {
    const nulls = await call('discover', { manifests: null })
    expect(nulls.result).toEqual({ projects: [] })
    const one = await call('discover', { manifests: ['a'] })
    expect(one.result.projects[0].name).toBe('a')
    const published = await call('publish', { project: { name: 'a', root: 'a', language: 'w', kind: 'lib' }, version: '1.0.0' })
    expect(published.result).toEqual({ published: true })
  })

  it('answers a method the plugin lacks with MethodNotFound, whether or not the protocol knows it', async () => {
    const known = await call('contribute', {})
    expect(known.error).toEqual({ code: -32601, message: 'method not supported: contribute' })
    const unknown = await call('nope', {})
    expect(unknown.error).toEqual({ code: -32601, message: 'method not supported: nope' })
  })

  it('turns what a method throws into PluginFailure with its message', async () => {
    const failed = await call('doctor', { projects: [] })
    expect(failed.error).toEqual({ code: -32000, message: 'the toolchain is missing: install node 24' })
  })

  it('answers params of the wrong shape with InvalidParams, and a result that cannot be written with InternalError', async () => {
    const wrong = await call('discover', { manifests: 'a' })
    expect(wrong.error.code).toBe(-32602)
    const bad = await call('generate', { kind: 'app', name: 'n', directory: 'd', workspace: { name: 'w' } })
    expect(bad.error.code).toBe(-32603)
    expect(bad.error.message).toContain('result.files must be an array')
  })

  it('answers shutdown with an empty object and stops reading', async () => {
    const result = answers(await talk(plugin, [initialize, { jsonrpc: '2.0', id: 2, method: 'shutdown', params: {} }, { jsonrpc: '2.0', id: 3, method: 'discover', params: {} }]))
    expect(result).toHaveLength(2)
    expect(result[1]).toEqual({ jsonrpc: '2.0', id: 2, result: {} })
  })

  it('answers no notification, and answers an unreadable body with ParseError and an id of null', async () => {
    const result = answers(await talk(plugin, [initialize, { jsonrpc: '2.0', method: 'discover', params: {} }, '{not json', '[]', { jsonrpc: '2.0', id: 4, method: 'shutdown' }]))
    expect(result.map(r => r.id)).toEqual([1, null, null, 4])
    expect(result[1].error.code).toBe(-32700)
    expect(result[2].error.code).toBe(-32700)
  })

  it('calls a method on the plugin object, so it can use this', async () => {
    class Counting {
      count = 0
      describe = describeWidget
      doctor () {
        this.count++

        return { findings: [{ check: `call ${this.count}`, ok: true }] }
      }
    }
    const result = answers(await talk(new Counting(), [initialize, { id: 2, method: 'doctor', params: {} }, { id: 3, method: 'doctor', params: {} }]))
    expect(result[2].result.findings[0].check).toBe('call 2')
  })
})

describe('serveStreams', () => {
  it('warns about half of a pair of methods, which is not advertised and not called', async () => {
    const plugin: SmithPlugin = { describe: describeWidget, install: () => ({}) }
    const conversation = await talk(plugin, [initialize, { id: 2, method: 'dependencies/install', params: {} }])
    expect(conversation.warnings).toHaveLength(1)
    expect(conversation.warnings[0]).toContain('install needs its pair')
    const result = answers(conversation)
    expect(result[0].result.capabilities).toEqual({})
    expect(result[1].error.code).toBe(-32601)
  })

  it('ends without an answer when the host closes the stream', async () => {
    const conversation = await talk({ describe: describeWidget }, [])
    expect(conversation.bodies).toEqual([])
  })

  it('rejects when the header is not a frame header', async () => {
    const input = new PassThrough()
    input.end('hello\r\n\r\n')
    await expect(serveStreams({ describe: describeWidget }, { input, output: new PassThrough() })).rejects.toThrow('malformed header')
  })
})
