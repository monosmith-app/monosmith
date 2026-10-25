import { describe, expect, it } from 'vitest'
import { RpcError } from '../protocol/index.ts'
import type { SmithPlugin } from '../serve/index.ts'
import { connectInProcess } from './connect-plugin.use-case.ts'

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
