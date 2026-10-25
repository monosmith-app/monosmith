import { spawnSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { converse, isAvailable } from './raw-conversation.fixture.ts'
import type { RawRequest } from './raw-conversation.fixture.ts'

const examples = resolve(import.meta.dirname, '..', '..', '..', 'docs', 'examples', 'plugins')
const go = isAvailable('go', ['version'])

let scratch = ''
let counter = 0

/** An empty directory for one conversation. */
function workspace (files: Record<string, string>): string {
  const dir = join(scratch, `ws${counter++}`)
  for (const [relative, content] of Object.entries(files)) {
    const full = join(dir, ...relative.split('/'))
    mkdirSync(dirname(full), { recursive: true })
    writeFileSync(full, content)
  }
  mkdirSync(dir, { recursive: true })

  return dir
}

/** Every file below a directory, by relative path. */
function snapshot (dir: string, prefix = ''): Record<string, string> {
  const files: Record<string, string> = {}
  const entries = readdirSync(join(dir, prefix), { withFileTypes: true })
  for (const entry of entries) {
    const relative = prefix === '' ? entry.name : `${prefix}/${entry.name}`
    if (entry.isDirectory()) Object.assign(files, snapshot(dir, relative))
    else files[relative] = readFileSync(join(dir, relative), 'utf8')
  }

  return files
}

function build (example: string): string {
  const out = join(scratch, `${example}${process.platform === 'win32' ? '.exe' : ''}`)
  const built = spawnSync('go', ['build', '-o', out, '.'], { cwd: resolve(examples, example), env: { ...process.env, GOWORK: 'off' }, encoding: 'utf8' })
  if (built.status !== 0) throw new Error(`go build ${example}: ${built.stderr}`)

  return out
}

interface Pair {
  goCommand: string[]
  tsCommand: string[]
  /** How each side writes the command of its own tasks, so it can be replaced by one name. */
  goSelf:    string
  tsSelf:    string
}

/** Runs the same requests against both plugins, each in its own copy of the workspace. */
async function both (pair: Pair, files: Record<string, string>, requests: (root: string) => RawRequest[], env: Record<string, string> = {}): Promise<{ go: string[], ts: string[], goRoot: string, tsRoot: string }> {
  const goRoot = workspace(files)
  const tsRoot = workspace(files)
  const goBodies = await converse(pair.goCommand, goRoot, requests(goRoot), env)
  const tsBodies = await converse(pair.tsCommand, tsRoot, requests(tsRoot), env)
  const go = goBodies.map(body => body.replaceAll(pair.goSelf, '"<self>"'))
  const ts = tsBodies.map(body => body.replaceAll(pair.tsSelf, '"<self>"'))

  return { go, ts, goRoot, tsRoot }
}

/** The Go program built from an example, and the TypeScript example that is the same plugin. */
function pairOf (goExample: string, tsExample: string): Pair {
  const goBinary = build(goExample)
  const tsPlugin = resolve(examples, tsExample, 'plugin.ts')

  return {
    goCommand: [goBinary],
    tsCommand: [process.execPath, tsPlugin],
    goSelf:    JSON.stringify(goBinary),
    tsSelf:    `"node",${JSON.stringify(tsPlugin)}`,
  }
}

/** The bodies in the order of their ids: a plugin that runs requests together may answer them in any order. */
function byId (bodies: string[]): string[] {
  return [...bodies].sort((a, b) => idOf(a) - idOf(b))
}

function idOf (body: string): number {
  return (JSON.parse(body) as { id: number }).id
}

/** The error code of each answer, or undefined for a result. */
function codes (bodies: string[]): unknown[] {
  return bodies.map(body => (JSON.parse(body) as { error?: { code: number } }).error?.code)
}

const initialize = (workspaceRoot: string): RawRequest => ({ method: 'initialize', params: { protocolVersions: [1], host: { name: 'smith', version: '1.62.0' }, workspaceRoot } })

const project = (name: string, root: string, kind: string): Record<string, string> => ({ name, root, language: 'widget', kind })

/**
 * The compatibility rule of the SDKs: for the same plugin logic, a TypeScript
 * plugin and a Go plugin give byte-identical responses. Each pair below is the
 * same plugin written twice; the one thing that cannot be equal is the command
 * of the tasks (the Go program's path, or `node` and the script), which is
 * replaced by a name on both sides.
 */
describe.skipIf(!go)('a TypeScript plugin and a Go plugin with the same logic', () => {
  beforeAll(() => {
    scratch = mkdtempSync(join(tmpdir(), 'sdk-cross-'))
  })

  afterAll(() => {
    rmSync(scratch, { recursive: true, force: true })
  })

  describe('language-widget-ts and language-widget', () => {
    const files = {
      'libs/core/widget.json': '{\n  "name": "core",\n  "version": "1.0.0",\n  "description": "Core <widgets> & more",\n  "kind": "lib"\n}\n',
      'apps/shop/widget.json': '{"name":"shop","version":"2.1.0","kind":"app","uses":["core","other"],"license":"MIT","stack":"widget-plain","description":"café"}',
      'libs/skip/widget.json': '{}',
    }
    it('answer the handshake, discover, doctor, version/write and generate with the same bytes', async () => {
      const pair = pairOf('language-widget', 'language-widget-ts')
      const manifests = ['libs/core/widget.json', 'apps/shop/widget.json', 'libs/skip/widget.json']
      const projects = [project('core', 'libs/core', 'lib'), project('shop', 'apps/shop', 'app')]
      const { go, ts, goRoot, tsRoot } = await both(pair, files, root => [
        initialize(root),
        { method: 'discover', params: { manifests } },
        { method: 'discover', params: { manifests: [] } },
        { method: 'doctor', params: { projects } },
        { method: 'doctor', params: { projects: [] } },
        { method: 'version/write', params: { project: projects[0], version: '1.1.0', dryRun: true } },
        { method: 'version/write', params: { project: projects[0], version: '1.1.0' } },
        { method: 'version/write', params: { project: projects[1], version: '2.2.0' } },
        { method: 'generate', params: { kind: 'package', name: 'gen', directory: 'packages/gen', workspace: { name: 'acme', scope: '@acme' }, stack: 'widget-plain' } },
        { method: 'generate', params: { kind: 'app', name: 'plain', directory: 'apps/plain', workspace: { name: 'acme' } } },
        { method: 'contribute', params: { projects } },
        { method: 'nope', params: {} },
      ])
      expect(ts).toHaveLength(go.length)
      expect(go.length).toBeGreaterThan(12)
      for (const [i, body] of go.entries()) expect(ts[i], `response ${i + 1}`).toBe(body)
      // The calls that write leave the same files, byte for byte.
      expect(snapshot(tsRoot)).toEqual(snapshot(goRoot))
      expect(Object.keys(snapshot(tsRoot))).toContain('packages/gen/widget.json')
      // Spot checks, so that equal does not mean equally empty.
      expect(go[0]).toBe('{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget","version":"0.1.0"},"kind":"language","capabilities":{"discover":true,"generate":true,"version":true,"doctor":true},"languages":["widget"],"manifests":["widget.json"],"roles":["language","scaffold"]}}')
      expect(go[1]).toContain(String.raw`Core \u003cwidgets\u003e \u0026 more`)
      expect(go[1]).toContain('"imports":["core","other"]')
    })

    it('refuse a request before initialize, a version they do not speak, and a method they lack with the same bytes', async () => {
      const pair = pairOf('language-widget', 'language-widget-ts')
      const early = await both(pair, files, () => [{ method: 'discover', params: { manifests: [] } }])
      expect(early.ts).toEqual(early.go)
      expect(early.go[0]).toBe('{"jsonrpc":"2.0","id":1,"error":{"code":-32600,"message":"initialize must come first"}}')

      const version = await both(pair, files, root => [{ method: 'initialize', params: { protocolVersions: [2, 3], host: { name: 'smith', version: '9' }, workspaceRoot: root } }])
      expect(version.ts).toEqual(version.go)
      expect(version.go[0]).toContain('no common protocol version: host speaks [2 3], plugin speaks 1')
    })

    it('ignore a cancel notification, which they did not advertise, with the same bytes', async () => {
      const { go, ts } = await both(pairOf('language-widget', 'language-widget-ts'), files, root => [
        initialize(root),
        { method: 'cancel', params: { id: 1 }, notification: true },
        { method: 'discover', params: { manifests: [] } },
      ])
      expect(ts).toEqual(go)
      // initialize, discover and shutdown are answered; the notification is not, and no capability or concurrency is declared.
      expect(go).toHaveLength(3)
      expect(go[0]).not.toContain('"cancel"')
      expect(go[0]).not.toContain('"concurrency"')
    })

    it('answer params of the wrong shape with the same code', async () => {
      const { go, ts } = await both(pairOf('language-widget', 'language-widget-ts'), files, root => [initialize(root), { method: 'discover', params: { manifests: 'libs/core/widget.json' } }, { method: 'doctor', params: [] }])
      expect(codes(ts)).toEqual(codes(go))
      expect(codes(go).slice(1, 3)).toEqual([-32602, -32602])
    })
  })

  describe('structure-scan-node and structure-scan-go, which are cancellable and declare a concurrency', () => {
    const files = Object.fromEntries([
      ...['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'].map(name => [`libs/core/${name}.txt`, name]),
      ['libs/core/data/big.bin', 'x'.repeat((1 << 20) + 1)],
    ])
    const projects = [project('core', 'libs/core', 'lib')]

    it('answer the handshake and doctor with the same bytes, concurrency and cancel included', async () => {
      const { go: goBodies, ts: tsBodies } = await both(pairOf('structure-scan-go', 'structure-scan-node'), files, root => [
        initialize(root),
        { method: 'doctor', params: { projects } },
        { method: 'doctor', params: { projects: [] } },
        { method: 'cancel', params: { id: 424_242 }, notification: true },
        { method: 'doctor', params: { projects }, thenWaitMs: 500 }, // shutdown cancels what is in flight, so let the scans finish first
      ])
      // Responses may come in any order (the scans run together), so they are compared by id.
      const go = byId(goBodies)
      const ts = byId(tsBodies)
      expect(ts).toHaveLength(go.length)
      expect(go).toHaveLength(5) // the notification gets no answer
      for (const [i, body] of go.entries()) expect(ts[i], `response ${i + 1}`).toBe(body)
      expect(go[0]).toBe('{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget-scan","version":"0.1.0"},"kind":"convention","capabilities":{"doctor":true,"cancel":true},"concurrency":4,"languages":["widget"],"roles":["structure"]}}')
      expect(go[1]).toContain('"check":"core has no file over 1 MiB","ok":false,"detail":"data/big.bin"')
    })

    it('stop a scan that is cancelled midway, and answer it with the cancelled error and the same bytes', async () => {
      const { go: goBodies, ts: tsBodies } = await both(pairOf('structure-scan-go', 'structure-scan-node'), files, root => [
        initialize(root),
        { method: 'doctor', params: { projects }, thenWaitMs: 250 },
        { method: 'cancel', params: { id: 2 }, notification: true, thenWaitMs: 150 },
        { method: 'doctor', params: { projects: [] } },
      ], { WIDGET_SCAN_DELAY_MS: '100' })
      // Two concurrent requests may be answered in either order; each frame must be equal.
      const go = byId(goBodies)
      const ts = byId(tsBodies)
      expect(ts).toEqual(go)
      expect(go).toHaveLength(4)
      expect(go[1]).toBe('{"jsonrpc":"2.0","id":2,"error":{"code":-32800,"message":"request cancelled: the scan was stopped"}}')
    })
  })

  describe('structure-layout-node and structure-layout-go', () => {
    const files = { 'libs/core/README.md': '# core\n' }
    it('answer the handshake, contribute and doctor with the same bytes', async () => {
      const projects = [project('core', 'libs/core', 'lib'), project('web', 'apps/web', 'app')]
      const { go, ts } = await both(pairOf('structure-layout-go', 'structure-layout-node'), files, root => [
        initialize(root),
        { method: 'contribute', params: { projects } },
        { method: 'contribute', params: { projects: [] } },
        { method: 'doctor', params: { projects } },
        { method: 'doctor', params: { projects: [] } },
        { method: 'discover', params: { manifests: [] } },
      ])
      expect(ts).toHaveLength(go.length)
      for (const [i, body] of go.entries()) expect(ts[i], `response ${i + 1}`).toBe(body)
      expect(go[0]).toBe('{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget-layout","version":"0.1.0"},"kind":"convention","capabilities":{"doctor":true,"contribute":true},"languages":["widget"],"roles":["structure"]}}')
      expect(go[1]).toContain('"dependsOn":{"lint":["layout-check"]}')
      expect(go[3]).toContain('"warning":true')
    })
  })
})
