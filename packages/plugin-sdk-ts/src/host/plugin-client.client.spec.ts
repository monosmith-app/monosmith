import { describe, expect, it } from 'vitest'
import { UNLIMITED_IN_FLIGHT } from '../concurrency/index.ts'
import { RpcError } from '../protocol/index.ts'
import type { SmithPlugin } from '../serve/index.ts'
import { CallInterruptedError } from './call-interrupted.error.ts'
import { connectInProcess } from './connect-plugin.use-case.ts'
import type { CallOptions } from './host-options.contract.ts'
import type { PluginClient } from './plugin-client.client.ts'

const widget: SmithPlugin = {
  describe: (params) => ({ plugin: { name: 'widget', version: '0.1.0' }, kind: 'language', manifests: ['widget.json'], languages: [params.host.name] }),
  discover: ({ manifests }) => {
    if (manifests.includes('bad/widget.json')) throw new Error('bad manifest')

    return { projects: manifests.map(manifest => ({ name: manifest, root: manifest, language: 'widget', kind: 'lib' as const })) }
  },
}

describe('PluginClient', () => {
  it('performs the handshake and tells what the plugin advertised', async () => {
    const client = await connectInProcess(widget, { workspaceRoot: '/ws', host: { name: 'tests', version: '1' } })
    expect(client.info.plugin).toEqual({ name: 'widget', version: '0.1.0' })
    expect(client.info.languages).toEqual(['tests'])
    expect(client.info.protocolVersion).toBe(1)
    expect(client.supports('discover')).toBe(true)
    expect(client.supports('publish')).toBe(false)
    expect(client.supports('dependencies/install')).toBe(false)
    expect(client.supports('initialize')).toBe(true)
    await client.close()
  })

  it('calls a method and decodes its result', async () => {
    const client = await connectInProcess(widget, { workspaceRoot: '/ws' })
    const result = await client.call('discover', { manifests: ['a/widget.json'] })
    expect(result.projects.map(p => p.name)).toEqual(['a/widget.json'])
    await client.close()
  })

  it('rejects with the code and message of an error answer, and stays usable', async () => {
    const client = await connectInProcess(widget, { workspaceRoot: '/ws' })
    const failing = expect(client.call('discover', { manifests: ['bad/widget.json'] })).rejects
    await failing.toBeInstanceOf(RpcError)
    await failing.toMatchObject({ code: -32000, message: 'bad manifest' })
    const missing = client.call('publish', { project: { name: 'a', root: 'a', language: 'w', kind: 'lib' }, version: '1.0.0' })
    await expect(missing).rejects.toMatchObject({ code: -32601 })
    await expect(client.call('discover', { manifests: [] })).resolves.toEqual({ projects: [] })
    await client.close()
  })

  it('serialises calls in the order they were made', async () => {
    const order: string[] = []
    const slow: SmithPlugin = {
      ...widget,
      discover: async ({ manifests }) => {
        await new Promise(resolve => setTimeout(resolve, manifests.length === 1 ? 30 : 0))
        order.push(manifests.join(','))

        return { projects: [] }
      },
    }
    const client = await connectInProcess(slow, { workspaceRoot: '/ws' })
    await Promise.all([client.call('discover', { manifests: ['first'] }), client.call('discover', { manifests: ['s', 'econd'] })])
    expect(order).toEqual(['first', 's,econd'])
    await client.close()
  })

  it('gives up on a plugin that does not answer in time, and says so', async () => {
    const silent: SmithPlugin = { ...widget, discover: () => new Promise(() => undefined) }
    const client = await connectInProcess(silent, { workspaceRoot: '/ws' })
    await expect(client.call('discover', { manifests: [] }, { timeoutMs: 40 })).rejects.toThrow('discover: plugin did not answer within 40ms')
  })

  it('reports a handshake that fails', async () => {
    const refusing: SmithPlugin = { describe: () => { throw new Error('cannot start') } }
    await expect(connectInProcess(refusing, { workspaceRoot: '/ws' })).rejects.toThrow('initializing plugin: cannot start')
  })
})

/** A plugin whose discover does what the manifest name says: `block*` waits for the release or the abort, `deaf*` waits for the release and ignores the abort, `hold*` takes a moment. */
function lab (options: { concurrency?: number, cancellable?: boolean } = {}) {
  const state = { active: 0, peak: 0, started: [] as string[], release: () => {}, aborts: 0 }
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
        if (name.startsWith('block')) {
          await new Promise<void>((resolve, reject) => {
            void released.then(resolve)
            signal.addEventListener('abort', () => {
              state.aborts++
              reject(new Error('stopped'))
            }, { once: true })
          })
        }
        if (name.startsWith('deaf')) await released
        if (name.startsWith('hold')) await new Promise(resolve => setTimeout(resolve, 20))
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

async function discover (client: PluginClient, name: string, options?: CallOptions): Promise<string> {
  const { projects } = await client.call('discover', { manifests: [name] }, options)

  return projects[0]?.name ?? ''
}

describe('PluginClient and the plugin\'s concurrency', () => {
  it('serialises the calls to a plugin that declares nothing', async () => {
    const { plugin, state } = lab()
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws' })
    await Promise.all(Array.from({ length: 8 }, async () => await discover(client, 'hold')))
    expect(state.peak).toBe(1)
    expect(client.peakInFlight).toBe(1)
    await client.close()
  })

  it('has as many calls in flight as the plugin declared, and no more', async () => {
    const { plugin, state } = lab({ concurrency: 3 })
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws' })
    await Promise.all(Array.from({ length: 12 }, async () => await discover(client, 'hold')))
    expect(state.peak).toBe(3)
    expect(client.peakInFlight).toBe(3)
    await client.close()
  })

  it('keeps to its own bound for a plugin that declares no limit', async () => {
    const { plugin } = lab({ concurrency: -1 })
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws' })
    await Promise.all(Array.from({ length: UNLIMITED_IN_FLIGHT * 2 }, async () => await discover(client, 'hold')))
    expect(client.peakInFlight).toBeGreaterThan(1)
    expect(client.peakInFlight).toBeLessThanOrEqual(UNLIMITED_IN_FLIGHT)
    await client.close()
  })

  it('never sends a call that was cancelled while it waited for a place', async () => {
    const { plugin, state } = lab()
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws' })
    const running = discover(client, 'block-a')
    await until(() => state.started.length === 1)
    const controller = new AbortController()
    const waiting = discover(client, 'block-b', { signal: controller.signal })
    controller.abort()
    await expect(waiting).rejects.toMatchObject({ name: 'CallInterruptedError', reason: 'cancelled', killed: false })
    state.release()
    await expect(running).resolves.toBe('block-a')
    expect(state.started).toEqual(['block-a'])
    await client.close()
  })
})

describe('PluginClient and cancelling a call', () => {
  it('asks a cancellable plugin to stop and rejects with its cancelled error, leaving it usable', async () => {
    const { plugin, state } = lab({ cancellable: true })
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws' })
    expect(client.supports('cancel')).toBe(true)
    const controller = new AbortController()
    const call = discover(client, 'block-a', { signal: controller.signal })
    await until(() => state.started.length === 1)
    controller.abort()
    let error: unknown
    try {
      await call
    } catch (error_) {
      error = error_
    }
    expect(error).toBeInstanceOf(CallInterruptedError)
    expect(error).toMatchObject({ reason: 'cancelled', killed: false, response: { code: -32800 } })
    expect(state.aborts).toBe(1)
    await expect(discover(client, 'after')).resolves.toBe('after')
    await client.close()
  })

  it('does the same when the time runs out', async () => {
    const { plugin } = lab({ cancellable: true })
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws' })
    await expect(discover(client, 'block-a', { timeoutMs: 60 })).rejects.toMatchObject({ reason: 'timeout', killed: false, response: { code: -32800 } })
    await expect(discover(client, 'after')).resolves.toBe('after')
    await client.close()
  })

  it('kills a plugin that ignores the cancel once the grace period is over', async () => {
    const { plugin, state } = lab({ cancellable: true })
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws', cancelGraceMs: 150 })
    const controller = new AbortController()
    const call = discover(client, 'deaf-a', { signal: controller.signal })
    await until(() => state.started.length === 1)
    const began = Date.now()
    controller.abort()
    await expect(call).rejects.toMatchObject({ reason: 'cancelled', killed: true })
    expect(Date.now() - began).toBeGreaterThanOrEqual(140)
    await expect(discover(client, 'after', { timeoutMs: 200 })).rejects.toThrow() // the plugin is gone
    state.release()
  })

  it('resolves with the result a plugin finishes while the host is stopping the call', async () => {
    const { plugin, state } = lab({ cancellable: true })
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws', cancelGraceMs: 2000 })
    const controller = new AbortController()
    const call = discover(client, 'deaf-a', { signal: controller.signal })
    await until(() => state.started.length === 1)
    controller.abort()
    setTimeout(() => { state.release() }, 50)
    await expect(call).resolves.toBe('deaf-a')
    await client.close()
  })

  it('never sends cancel to a plugin that did not advertise it', async () => {
    const { plugin, state } = lab()
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws', cancelGraceMs: 100 })
    expect(client.supports('cancel')).toBe(false)
    const controller = new AbortController()
    const call = discover(client, 'block-a', { signal: controller.signal })
    await until(() => state.started.length === 1)
    controller.abort()
    await expect(call).rejects.toMatchObject({ reason: 'cancelled', killed: true })
    expect(state.aborts).toBe(0)
    state.release()
  })

  it('kills a plugin that cannot be cancelled at once when the time runs out', async () => {
    const { plugin, state } = lab()
    const client = await connectInProcess(plugin, { workspaceRoot: '/ws', cancelGraceMs: 60_000 })
    const began = Date.now()
    await expect(discover(client, 'block-a', { timeoutMs: 40 })).rejects.toThrow('discover: plugin did not answer within 40ms')
    expect(Date.now() - began).toBeLessThan(2000)
    state.release()
  })
})
