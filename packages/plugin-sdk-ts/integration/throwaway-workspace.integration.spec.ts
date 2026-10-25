import { spawnSync } from 'node:child_process'
import { appendFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { JsonDocument } from '../src/jsondoc/index.ts'
import { isAvailable } from './raw-conversation.fixture.ts'

const examples = resolve(import.meta.dirname, '..', '..', '..', 'docs', 'examples', 'plugins')
const smith = isAvailable('smith', ['--version'])

/** A path as a TOML string in smith.toml: forward slashes, so Windows paths need no escaping. */
const toml = (path: string): string => JSON.stringify(path.replaceAll('\\', '/'))

let root = ''
let workspace = ''

/** Runs smith in the workspace, without the variable that would send a session elsewhere. */
function run (...args: string[]): { status: number | null, out: string } {
  const env: NodeJS.ProcessEnv = { ...process.env, NO_COLOR: '1' }
  delete env.SMITH_WORKSPACE_ROOT
  const result = spawnSync('smith', args, { cwd: workspace === '' ? root : workspace, encoding: 'utf8', env })

  return { status: result.status, out: `${result.stdout}${result.stderr}` }
}

// A real workspace that uses the two TypeScript examples as external plugins.
describe.skipIf(!smith)('a throwaway workspace using the TypeScript example plugins', () => {
  beforeAll(() => {
    root = mkdtempSync(join(tmpdir(), 'sdk-workspace-'))
    expect(run('new', 'ws').status).toBe(0)
    workspace = join(root, 'ws')
    appendFileSync(join(workspace, 'smith.toml'), [
      '',
      '[plugins.external.widget]',
      `command = [${toml(process.execPath)}, ${toml(resolve(examples, 'language-widget-ts', 'plugin.ts'))}]`,
      '',
      '[plugins.external.widget-layout]',
      `command = [${toml(process.execPath)}, ${toml(resolve(examples, 'structure-layout-node', 'plugin.ts'))}]`,
      '',
      '[conventions.widget-layout]',
      '',
    ].join('\n'))
  })

  afterAll(() => {
    rmSync(root, { recursive: true, force: true })
  })

  it('creates widget projects with the plugin, and discovers them with their graph edge', () => {
    expect(run('add', 'lib', 'core', '--lang', 'widget').status).toBe(0)
    expect(run('add', 'app', 'shop', '--lang', 'widget').status).toBe(0)
    const manifest = join(workspace, 'apps', 'shop', 'widget.json')
    const document = JsonDocument.parse(readFileSync(manifest, 'utf8'))
    document.setJson('uses', ['core'])
    writeFileSync(manifest, document.render())

    const projects = run('projects')
    expect(projects.out).toMatch(/^shop .*app .*widget/m)
    expect(projects.out).toMatch(/^core .*lib .*widget/m)
    expect(run('graph').out).toContain('core')
  })

  it('runs the build and test tasks, and serves the second run from the cache', () => {
    const first = run('run', 'build', 'test')
    expect(first.out).toContain('tasks succeeded')
    expect(first.status).toBe(0)
    expect(readFileSync(join(workspace, 'dist', 'libs', 'core', 'widget.txt'), 'utf8')).toBe('core 0.0.0\n')
    const second = run('run', 'build', 'test')
    expect(second.out).toContain('from cache')
  })

  it('adds the layout-check task of the structure plugin to every widget project, and fails on a junk drawer', () => {
    const clean = run('run', 'layout-check')
    expect(clean.out).toContain('layout-check')
    expect(clean.status).toBe(0)
    // A file in it: the task is cached on the files of the project, and an empty folder is none.
    mkdirSync(join(workspace, 'libs', 'core', 'utils'))
    writeFileSync(join(workspace, 'libs', 'core', 'utils', 'index.txt'), 'everything')
    const dirty = run('run', 'layout-check')
    expect(dirty.status).not.toBe(0)
    expect(dirty.out).toContain('folders named for nothing: utils')
  })

  it('makes the structure plugin\'s doctor finding a warning about the README it asks for', () => {
    const doctor = run('doctor')
    expect(doctor.out).toContain('has a README.md')
  })
})
