import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, posix } from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { connectInProcess } from '../host/index.ts'
import type { SmithPlugin } from '../serve/index.ts'
import { checkPlugin } from './check-plugin.use-case.ts'
import type { ConformanceReport } from './conformance-report.model.ts'

let ws: string
beforeEach(() => { ws = mkdtempSync(join(tmpdir(), 'sdk-conformance-')) })
afterEach(() => { rmSync(ws, { recursive: true, force: true }) })

// The three plugins of the Go SDK's conformance tests: one that keeps its promises, and two that break several.

/** A language whose manifest, fig.txt, holds "kind stack". */
function good (): SmithPlugin {
  return {
    describe: () => ({ plugin: { name: 'fig', version: '1.0.0' }, kind: 'language', manifests: ['fig.txt'], stacks: [{ name: 'fig-plain', language: 'fig' }] }),
    generate: (params) => {
      mkdirSync(join(ws, ...params.directory.split('/')), { recursive: true })
      writeFileSync(join(ws, ...params.directory.split('/'), 'fig.txt'), `${params.kind} ${params.stack ?? ''}`)

      return { files: [`${params.directory}/fig.txt`] }
    },
    discover: ({ manifests }) => ({
      projects: manifests.map((manifest) => {
        const [kind, stack] = readFileSync(join(ws, ...manifest.split('/')), 'utf8').split(' ', 2)
        const root = posix.dirname(manifest)

        return { name: posix.basename(root), root, language: 'fig', kind: kind as 'app', stack, tasks: { build: { command: ['fig', 'build'] } } }
      }),
    }),
    writeVersion: () => ({ files: [] }),
  }
}

/** The good language, able to change kinds; `how` says how it misbehaves, if it does. */
function changer (how: '' | 'nothing' | 'missing' | 'write'): SmithPlugin {
  const plugin = good()
  plugin.changeKind = ({ project, to }) => {
    if (how === 'nothing') return {}
    if (how === 'missing') return { files: [{ path: `${project.root}/nowhere.txt`, content: 'x' }] }
    const manifest = `${project.root}/fig.txt`
    if (how === 'write') writeFileSync(join(ws, ...manifest.split('/')), `${to} conformance`)

    return { files: [{ path: manifest, content: `${to} conformance` }] }
  }

  return plugin
}

async function reportOf (plugin: SmithPlugin): Promise<ConformanceReport> {
  const client = await connectInProcess(plugin, { workspaceRoot: ws })
  const report = await checkPlugin(client, { workspace: ws })
  await client.close()

  return report
}

/** The failures of the static checks of a plugin served in this process. */
async function failuresOf (plugin: SmithPlugin): Promise<string> {
  const client = await connectInProcess(plugin, { workspaceRoot: ws })
  const report = await checkPlugin(client)

  return report.failures()
}

describe('checkPlugin', () => {
  it('passes a plugin that keeps its promises, and runs the generation checks when asked', async () => {
    const client = await connectInProcess(good(), { workspaceRoot: ws })
    const report = await checkPlugin(client, { workspace: ws })
    expect(report.failures()).toBe('')
    expect(report.ok).toBe(true)
    const checks = report.results.map(result => result.check)
    expect(checks).toEqual(expect.arrayContaining([
      'names itself', 'has a known kind', 'plays a known role: language', 'plays a known role: toolchain', 'plays a known role: scaffold',
      'a language plugin discovers projects', 'a scaffold generates projects', 'stack fig-plain is complete', 'discover answers with no manifests',
      'generates an app', 'generates a lib', 'generates a package', 'discovers the generated package', 'the generated app records its stack',
      "the generated lib's tasks have commands", "writes the app's version (dry run)",
    ]))
    expect(JSON.stringify(report)).toMatch(/^\{"plugin":\{"name":"fig","version":"1.0.0"\},"results":\[\{"check":"names itself","ok":true\}/)
    await client.close()
  })

  it('generates only the kinds a plugin declares', async () => {
    const plugin = good()
    const describe = plugin.describe
    plugin.describe = async params => ({ ...await describe(params), kinds: ['app'] })
    const report = await reportOf(plugin)
    expect(report.failures()).toBe('')
    const checks = report.results.map(result => result.check)
    expect(checks).toContain('discovers the generated app')
    expect(checks).not.toContain('generates a lib')
    expect(checks).not.toContain('generates a package')
  })

  describe('a plugin that changes kinds', () => {
    it('is asked what each generated kind becomes, and passes', async () => {
      const report = await reportOf(changer(''))
      expect(report.failures()).toBe('')
      expect(report.results.map(result => result.check)).toEqual(expect.arrayContaining([
        'answers what making the app a lib edits (dry run)', 'answers what making the lib a package edits (dry run)', 'answers what making the package a lib edits (dry run)',
      ]))
    })

    it('is not asked when it does not declare changeKind', async () => {
      const report = await reportOf(good())
      expect(report.results.some(result => result.check.includes('edits (dry run)'))).toBe(false)
    })

    it.each([
      ['nothing', 'neither a refusal nor an edit'],
      ['missing', 'files that do not exist'],
      ['write', 'it wrote to the workspace'],
    ] as const)('reports an answer that is %s', async (how, want) => {
      const report = await reportOf(changer(how))
      expect(report.failures()).toContain(want)
    })
  })

  it('does not generate without a workspace', async () => {
    const client = await connectInProcess(good(), { workspaceRoot: ws })
    const report = await checkPlugin(client)
    expect(report.results.some(result => result.check.startsWith('generates'))).toBe(false)
    await client.close()
  })

  it('reports a plugin that does nothing, provides a task without the role, and publishes an incomplete stack', async () => {
    const bad: SmithPlugin = { describe: () => ({ plugin: { name: 'bad', version: '' }, kind: 'convention', provides: ['test'], stacks: [{ name: 'half', language: '' }] }) }
    const failures = await failuresOf(bad)
    for (const want of ['names itself', 'a structure plugin does something', 'only a toolchain provides tasks', 'stack half is complete']) {
      expect(failures, want).toContain(want)
    }
    expect(failures).toContain('it provides [test] without the toolchain role')
  })

  it('checks the levels and references a plugin declares', async () => {
    const refs: SmithPlugin = {
      describe: () => ({
        plugin:    { name: 'refs', version: '1.0.0' },
        kind:      'convention',
        roles:     ['toolchain'],
        requires:  [{}, { level: 'platform' as never, name: 'x' }],
        conflicts: [{ name: 'refs' }],
      }),
      contribute: () => ({}),
    }
    const failures = await failuresOf(refs)
    for (const want of ['a toolchain names the slots it fills', 'requires plugin: a known reference', 'requires platform plugin x: a known reference', 'conflicts with plugin refs: a known reference']) {
      expect(failures, want).toContain(want)
    }
  })

  it('reports a method that fails, with the plugin\'s message', async () => {
    const plugin: SmithPlugin = { describe: () => ({ plugin: { name: 'x', version: '1' }, kind: 'convention' }), doctor: () => { throw new Error('toolchain missing') } }
    expect(await failuresOf(plugin)).toContain('doctor answers with no projects: toolchain missing')
  })

  it('reports a generated project that discovery does not find, or finds with the wrong stack', async () => {
    const sloppy = good()
    sloppy.discover = ({ manifests }) => ({ projects: manifests.slice(0, 1).map(manifest => ({ name: 'x', root: posix.dirname(manifest), language: 'fig', kind: 'lib' as const, tasks: { t: { command: [] } } })) })
    const client = await connectInProcess(sloppy, { workspaceRoot: ws })
    const report = await checkPlugin(client, { workspace: ws })
    const failures = report.failures()
    expect(failures).toContain('discovers the generated lib: no project at libs/clib among 1 discovered')
    expect(failures).toContain('the generated app keeps its kind: discovered as "lib"')
    expect(failures).toContain('the generated app records its stack: discovered with stack ""')
    expect(failures).toContain("the generated app's tasks have commands: tasks without a command: [t]")
  })
})
