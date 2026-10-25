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

/** The host's end of a serve loop running in this process: requests go in as they are written, answers come out as they are written. */
function live (plugin: SmithPlugin) {
  const input = new PassThrough()
  const output = new PassThrough()
  const reader = new FrameReader()
  const received: Array<Record<string, any>> = []
  const waiting: Array<() => void> = []
  output.on('data', (chunk: Buffer) => {
    for (const body of reader.push(chunk)) received.push(JSON.parse(body) as Record<string, any>)
    for (const wake of waiting.splice(0)) wake()
  })
  const served = serveStreams(plugin, { input, output, log: () => {} })
  const host = {
    served,
    send (message: unknown) { input.write(frameMessage(JSON.stringify(message))) },
    request (id: number, method: string, params: unknown = {}) { host.send({ jsonrpc: '2.0', id, method, params }) },
    notify (method: string, params: unknown) { host.send({ jsonrpc: '2.0', method, params }) },
    discover (id: number, manifest: string) { host.request(id, 'discover', { manifests: [manifest] }) },
    /** The next n answers that have not been taken yet. */
    async next (n = 1): Promise<Array<Record<string, any>>> {
      const deadline = Date.now() + 5000
      while (received.length < n) {
        if (Date.now() > deadline) throw new Error(`only ${received.length} of ${n} answers`)
        await new Promise<void>((resolve) => {
          waiting.push(resolve)
          setTimeout(resolve, 20)
        })
      }

      return received.splice(0, n)
    },
    /** The next answer that has not been taken yet. */
    async one (): Promise<Record<string, any>> {
      const [answer] = await host.next()

      return answer
    },
    async silent (ms = 50): Promise<void> {
      await new Promise(resolve => setTimeout(resolve, ms))
      expect(received.splice(0)).toEqual([])
    },
    async initialize (): Promise<Record<string, any>> {
      host.send(initialize)

      const [answer] = await host.next()

      return answer
    },
    end () { input.end() },
  }

  return host
}

/** A plugin whose discover does what the manifest name says: `block*` waits to be released (or aborted), `stubborn*` waits and ignores the abort, `fail*` throws, anything else answers at once. */
function lab (options: { concurrency?: number, cancellable?: boolean } = {}) {
  const state = { active: 0, peak: 0, started: [] as string[], release: () => {}, aborted: 0 }
  const released = new Promise<void>((resolve) => { state.release = resolve })
  const plugin: SmithPlugin = {
    ...(options.cancellable === true) && { cancellable: true as const },
    describe: () => ({ plugin: { name: 'lab', version: '1.0.0' }, kind: 'language', manifests: ['lab.json'], ...options.concurrency !== undefined && { concurrency: options.concurrency } }),
    discover: async ({ manifests }, signal) => {
      const name = manifests[0] ?? ''
      state.active++
      state.peak = Math.max(state.peak, state.active)
      state.started.push(name)
      try {
        if (name.startsWith('fail')) throw new Error('the manifest is broken')
        if (name.startsWith('block')) {
          await new Promise<void>((resolve, reject) => {
            void released.then(resolve)
            signal.addEventListener('abort', () => {
              state.aborted++
              reject(new Error('aborted'))
            }, { once: true })
          })
        }
        if (name.startsWith('stubborn')) await released
        if (name.startsWith('slow')) await new Promise(resolve => setTimeout(resolve, Math.floor(Math.random() * 3)))
      } finally {
        state.active--
      }

      return { projects: [{ name, root: name, language: 'lab', kind: 'lib' as const }] }
    },
  }

  return { plugin, state }
}

async function until (condition: () => boolean): Promise<void> {
  const deadline = Date.now() + 5000
  while (!condition()) {
    if (Date.now() > deadline) throw new Error('the condition did not hold within 5 seconds')
    await new Promise(resolve => setTimeout(resolve, 5))
  }
}

describe('concurrency', () => {
  it('handles one request at a time, in order, for a plugin that declares nothing', async () => {
    const { plugin, state } = lab()
    const host = live(plugin)
    const { result } = await host.initialize()
    expect(result.concurrency).toBeUndefined()
    expect(result.capabilities.cancel).toBeUndefined()
    for (let id = 1; id <= 6; id++) host.discover(id, `slow-${id}`)
    const answered = await host.next(6)
    expect(answered.map(a => a.id)).toEqual([1, 2, 3, 4, 5, 6])
    expect(state.peak).toBe(1)
    host.end()
    await host.served
  })

  it('runs as many handlers together as the plugin declared, and no more', async () => {
    const { plugin, state } = lab({ concurrency: 3 })
    const host = live(plugin)
    const { result } = await host.initialize()
    expect(result.concurrency).toBe(3)
    for (let id = 1; id <= 8; id++) host.discover(id, `block-${id}`)
    await until(() => state.started.length === 3)
    await host.silent(50)
    expect(state.peak).toBe(3)
    expect(state.started).toHaveLength(3)
    state.release()
    const answered = await host.next(8)
    expect(new Set(answered.map(a => a.id)).size).toBe(8)
    expect(state.peak).toBe(3)
    host.end()
    await host.served
  })

  it('may answer out of order, each answer carrying the id of its request', async () => {
    const { plugin, state } = lab({ concurrency: 2 })
    const host = live(plugin)
    await host.initialize()
    host.discover(1, 'block-first')
    host.discover(2, 'quick')
    expect(await host.one()).toMatchObject({ id: 2, result: { projects: [{ name: 'quick' }] } })
    state.release()
    expect(await host.one()).toMatchObject({ id: 1 })
    host.end()
    await host.served
  })

  it('refuses a request id that is already in flight and a host that floods it', async () => {
    const { plugin, state } = lab({ concurrency: 1 })
    const host = live(plugin)
    await host.initialize()
    host.discover(1, 'block-a')
    await until(() => state.started.length === 1)
    host.discover(1, 'block-again')
    const { error } = await host.one()
    expect(error.code).toBe(-32600)
    for (let id = 2; id < 2 + 1024 + 5; id++) host.discover(id, 'block-queued')
    const refused = await host.next(5)
    for (const answer of refused) expect(answer.error).toMatchObject({ code: -32600, message: expect.stringContaining('too many requests') as string })
    state.release()
    expect(await host.next(1 + 1024)).toHaveLength(1025)
    host.end()
    await host.served
  })
})

describe('cancel', () => {
  it('is advertised by a plugin that says it is cancellable', async () => {
    const host = live(lab({ cancellable: true }).plugin)
    const { result } = await host.initialize()
    expect(result.capabilities.cancel).toBe(true)
    host.end()
    await host.served
  })

  it('aborts the signal of a running request and answers it with the cancelled error', async () => {
    const { plugin, state } = lab({ cancellable: true })
    const host = live(plugin)
    await host.initialize()
    host.discover(1, 'block-a')
    await until(() => state.started.length === 1)
    host.notify('cancel', { id: 1 })
    expect(await host.one()).toEqual({ jsonrpc: '2.0', id: 1, error: { code: -32800, message: 'request cancelled: aborted' } })
    host.discover(2, 'quick')
    const { result } = await host.one()
    expect(result.projects[0].name).toBe('quick')
    host.end()
    await host.served
  })

  it('ignores a cancel for an unknown or finished request, or with bad params', async () => {
    const host = live(lab({ cancellable: true }).plugin)
    await host.initialize()
    host.discover(1, 'quick')
    await host.next()
    host.notify('cancel', { id: 1 })
    host.notify('cancel', { id: 424_242 })
    host.notify('cancel', 'not params')
    await host.silent()
    host.discover(2, 'quick')
    const { id } = await host.one()
    expect(id).toBe(2)
    host.end()
    await host.served
  })

  it('answers with the result a request produced despite the cancel', async () => {
    const { plugin, state } = lab({ cancellable: true })
    const host = live(plugin)
    await host.initialize()
    host.discover(1, 'stubborn')
    await until(() => state.started.length === 1)
    host.notify('cancel', { id: 1 })
    await host.silent()
    state.release()
    const { result } = await host.one()
    expect(result.projects[0].name).toBe('stubborn')
    host.end()
    await host.served
  })

  it('keeps the plugin failure code for an error nobody asked for', async () => {
    const host = live(lab({ cancellable: true }).plugin)
    await host.initialize()
    host.discover(1, 'fail-by-itself')
    const { error } = await host.one()
    expect(error).toEqual({ code: -32000, message: 'the manifest is broken' })
    host.end()
    await host.served
  })

  it('answers a request that had not started at once, and never starts it', async () => {
    const { plugin, state } = lab({ cancellable: true, concurrency: 1 })
    const host = live(plugin)
    await host.initialize()
    host.discover(1, 'block-running')
    host.discover(2, 'block-waiting')
    await until(() => state.started.length === 1)
    host.notify('cancel', { id: 2 })
    expect(await host.one()).toMatchObject({ id: 2, error: { code: -32800 } })
    state.release()
    expect(await host.one()).toMatchObject({ id: 1 })
    expect(state.started).toEqual(['block-running'])
    host.end()
    await host.served
  })

  it('is ignored by a plugin that did not say it is cancellable, and its signal never aborts', async () => {
    const { plugin, state } = lab()
    const host = live(plugin)
    await host.initialize()
    host.discover(1, 'stubborn')
    await until(() => state.started.length === 1)
    host.notify('cancel', { id: 1 })
    await host.silent()
    state.release()
    const { result } = await host.one()
    expect(result.projects[0].name).toBe('stubborn')
    host.end()
    await host.served
  })

  it('is applied to what is in flight when the host sends shutdown, which is then answered', async () => {
    const { plugin, state } = lab({ cancellable: true })
    const host = live(plugin)
    await host.initialize()
    host.discover(1, 'block-a')
    await until(() => state.started.length === 1)
    host.request(2, 'shutdown')
    const answers = await host.next(2)
    expect(answers[0]).toMatchObject({ id: 1, error: { code: -32800 } })
    expect(answers[1]).toEqual({ jsonrpc: '2.0', id: 2, result: {} })
    await host.served
  })

  it('lets a plugin that cannot be cancelled finish before shutdown is answered', async () => {
    const { plugin, state } = lab()
    const host = live(plugin)
    await host.initialize()
    host.discover(1, 'stubborn')
    await until(() => state.started.length === 1)
    host.request(2, 'shutdown')
    await host.silent()
    state.release()
    const answers = await host.next(2)
    expect(answers.map(a => a.id)).toEqual([1, 2])
    await host.served
  })

  it('answers each of a stress of concurrent and cancelled requests exactly once, within the pool\'s bound', async () => {
    const total = 400
    const { plugin, state } = lab({ cancellable: true, concurrency: 8 })
    const host = live(plugin)
    await host.initialize()
    for (let id = 1; id <= total; id++) {
      host.discover(id, id % 3 === 0 ? 'block-x' : 'slow-x')
      if (id % 6 === 0) setTimeout(() => { host.notify('cancel', { id }) }, Math.floor(Math.random() * 5))
    }
    setTimeout(() => { state.release() }, 200)
    const answers = await host.next(total)
    const counts = new Map<number, number>()
    for (const answer of answers) {
      counts.set(answer.id as number, (counts.get(answer.id as number) ?? 0) + 1)
      if (answer.error !== undefined) expect(answer.error.code).toBe(-32800)
    }
    expect(counts.size).toBe(total)
    for (const count of counts.values()) expect(count).toBe(1)
    await host.silent(30)
    expect(state.peak).toBeLessThanOrEqual(8)
    host.end()
    await host.served
  })
})
