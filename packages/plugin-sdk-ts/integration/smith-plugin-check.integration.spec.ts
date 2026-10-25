import { spawnSync } from 'node:child_process'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import { isAvailable } from './raw-conversation.fixture.ts'

const examples = resolve(import.meta.dirname, '..', '..', '..', 'docs', 'examples', 'plugins')
const smith = isAvailable('smith', ['--version'])

/** `smith plugin check <node> <plugin.ts>`: the real conformance suite against a TypeScript example, started the way smith starts any plugin. */
function check (example: string, ...flags: string[]): { status: number | null, out: string } {
  return checkScript(example, 'plugin.ts', ...flags)
}

/** The same for any script of an example. */
function checkScript (example: string, script: string, ...flags: string[]): { status: number | null, out: string } {
  const env = { ...process.env }
  delete env.SMITH_WORKSPACE_ROOT
  const result = spawnSync('smith', ['plugin', 'check', process.execPath, resolve(examples, example, script), ...flags], { encoding: 'utf8', env })

  return { status: result.status, out: `${result.stdout}${result.stderr}` }
}

// These need the smith on the path. CI builds it from the commit before it runs the tests.
describe.skipIf(!smith)('smith plugin check, on the TypeScript examples', () => {
  it('passes the structure plugin that uses the SDK (structure-layout-node)', () => {
    const { status, out } = check('structure-layout-node')
    expect(out).toContain('widget-layout 0.1.0')
    expect(out).not.toContain('✖')
    expect(out).toContain('✔ a structure plugin does something')
    expect(out).toContain('✔ shuts down when asked')
    expect(status).toBe(0)
  })

  it('passes the language plugin (language-widget-ts), generating projects and discovering them again', () => {
    const { status, out } = check('language-widget-ts', '--generate')
    expect(out).toContain('widget 0.1.0')
    expect(out).not.toContain('✖')
    for (const kind of ['app', 'lib', 'package']) expect(out).toContain(`✔ discovers the generated ${kind}`)
    expect(status).toBe(0)
  })

  it('passes the plugin that is cancellable and declares a concurrency, and checks both (structure-scan-node)', () => {
    const { status, out } = check('structure-scan-node')
    expect(out).toContain('widget-scan 0.1.0')
    expect(out).not.toContain('✖')
    expect(out).toContain('✔ accepts 4 requests at the same time')
    expect(out).toContain('✔ answers a cancelled request with its result or the cancelled error')
    expect(out).toContain('✔ ignores a cancel for a request it does not know')
    expect(status).toBe(0)
  })

  it('still passes a plugin written before cancel and concurrency existed: it is called one request at a time and is never sent cancel', () => {
    // examples/node is the protocol by hand, with a strictly serial loop that knows nothing of either.
    const { status, out } = checkScript('node', 'plugin.mjs')
    expect(out).not.toContain('✖')
    expect(out).toContain('✔ declares a concurrency the protocol defines')
    expect(out).toContain('✔ answers 3 requests sent together, one at a time')
    expect(out).not.toContain('cancel')
    expect(status).toBe(0)
  })

  it('fails a plugin that breaks its promises, as the Go suite does', () => {
    const result = spawnSync('smith', ['plugin', 'check', process.execPath, '-e', 'process.exit(3)'], { encoding: 'utf8' })
    expect(result.status).not.toBe(0)
  })
})
