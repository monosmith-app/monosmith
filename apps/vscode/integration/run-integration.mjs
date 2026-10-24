/* eslint-disable unicorn/no-process-exit -- a command line program: its exit code is the result of the tests */
// Runs the integration tests: starts VS Code (a downloaded one, or the one
// VSCODE_EXECUTABLE_PATH names) with the extension loaded from dist, once per
// scenario, each on a fresh copy of a small workspace and with a fake `smith`
// that speaks `smith serve`, so no toolchain is needed.
//
//   npm run test:integration -w monosmith-vscode
//
// On Linux without a display run it under xvfb: xvfb-run -a npm run test:integration
import { spawn, spawnSync } from 'node:child_process'
import { chmodSync, copyFileSync, existsSync, mkdirSync, mkdtempSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { delimiter, dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { build } from 'esbuild'
import { downloadAndUnzipVSCode } from '@vscode/test-electron'

const here = dirname(fileURLToPath(import.meta.url))
const extensionRoot = dirname(here)
const bundled = join(extensionRoot, '..', '..', 'dist', 'integration')
const windows = process.platform === 'win32'

/** Where to find smith in each scenario, and whether the workspace is trusted. */
const SCENARIOS = [
  { name: 'ready', trusted: true, smith: 'setting' },
  { name: 'missing', trusted: true, smith: 'none' },
  { name: 'untrusted', trusted: false, smith: 'path' },
]

function log (message) {
  process.stdout.write(`[integration] ${message}\n`)
}

/** The fake smith as an executable file in a directory of its own. */
function installFakeSmith (dir) {
  mkdirSync(dir, { recursive: true })
  copyFileSync(join(here, 'fake-smith', 'fake-smith.cjs'), join(dir, 'fake-smith.cjs'))
  const executable = join(dir, windows ? 'smith.cmd' : 'smith')
  writeFileSync(executable, windows ? '@node "%~dp0fake-smith.cjs" %*\r\n' : '#!/bin/sh\nexec node "$(dirname "$0")/fake-smith.cjs" "$@"\n')
  if (!windows) chmodSync(executable, 0o755)

  return executable
}

function localCode () {
  if (process.env.VSCODE_EXECUTABLE_PATH) return process.env.VSCODE_EXECUTABLE_PATH
  if (!windows) return
  const found = spawnSync('where', ['code'], { encoding: 'utf8' }).stdout?.split(/\r?\n/, 1)[0]
  const exe = found ? join(dirname(dirname(found)), 'Code.exe') : undefined

  return exe !== undefined && existsSync(exe) ? exe : undefined
}

async function bundleSuite (name) {
  const dir = join(here, 'suites', name)
  const entries = readdirSync(dir).filter((file) => file.endsWith('.integration.spec.ts')).map((file) => join(dir, file))
  await build({
    entryPoints:  entries,
    outdir:       join(bundled, name),
    bundle:       true,
    platform:     'node',
    format:       'cjs',
    target:       'node20',
    external:     ['vscode'],
    outExtension: { '.js': '.js' },
    logLevel:     'warning',
  })

  return join(bundled, name)
}

async function runScenario (scenario, executable, work) {
  const suiteDir = await bundleSuite(scenario.name)
  const workspace = join(work, `${scenario.name}-workspace`)
  mkdirSync(join(workspace, '.vscode'), { recursive: true })
  writeFileSync(join(workspace, 'smith.toml'), 'schema = 1\n\n[workspace]\nname = "demo"\n')
  const smithDir = join(work, 'fake-bin')
  const fake = installFakeSmith(smithDir)
  if (scenario.smith === 'setting') writeFileSync(join(workspace, '.vscode', 'settings.json'), JSON.stringify({ 'monosmith.path': fake }))

  const system = windows ? [join(process.env.SystemRoot ?? String.raw`C:\Windows`, 'System32')] : ['/usr/bin', '/bin']
  const pathDirs = scenario.smith === 'path' ? [smithDir, dirname(process.execPath), ...system] : (scenario.smith === 'none' ? system : (process.env.PATH ?? '').split(delimiter))
  const userData = join(work, `${scenario.name}-user`)
  mkdirSync(join(userData, 'User'), { recursive: true })
  writeFileSync(join(userData, 'User', 'settings.json'), JSON.stringify({
    'update.mode':                            'none',
    'telemetry.telemetryLevel':               'off',
    'workbench.startupEditor':                'none',
    'security.workspace.trust.enabled':       !scenario.trusted,
    'security.workspace.trust.startupPrompt': 'never',
    'security.workspace.trust.banner':        'never',
  }))
  const args = [
    '--no-sandbox', '--disable-gpu-sandbox', '--disable-updates', '--skip-welcome', '--skip-release-notes', '--no-cached-data',
    '--disable-extensions',
    `--user-data-dir=${userData}`,
    `--extensions-dir=${join(work, `${scenario.name}-extensions`)}`,
    `--extensionDevelopmentPath=${extensionRoot}`,
    `--extensionTestsPath=${join(here, 'suite-runner.cjs')}`,
    ...(scenario.trusted ? ['--disable-workspace-trust'] : []),
    workspace,
  ]
  const env = { ...process.env, PATH: pathDirs.join(delimiter), MONOSMITH_SUITE_DIR: suiteDir }
  delete env.ELECTRON_RUN_AS_NODE
  log(`${scenario.name}: starting VS Code`)

  return await new Promise((resolve) => {
    const child = spawn(executable, args, { env, stdio: 'inherit' })
    const timer = setTimeout(() => {
      log(`${scenario.name}: timed out`)
      child.kill()
    }, 5 * 60_000)
    child.on('close', (code) => {
      clearTimeout(timer)
      resolve(code === 0)
    })
    child.on('error', (error) => {
      clearTimeout(timer)
      log(`${scenario.name}: ${error.message}`)
      resolve(false)
    })
  })
}

const built = spawnSync('npm', ['run', 'build'], { cwd: extensionRoot, stdio: 'inherit', shell: windows })
if (built.status !== 0) process.exit(built.status ?? 1)

const executable = localCode() ?? await downloadAndUnzipVSCode({ version: process.env.VSCODE_VERSION ?? 'stable', cachePath: join(extensionRoot, '.vscode-test') })
const work = mkdtempSync(join(tmpdir(), 'monosmith-integration-'))
const wanted = process.argv.slice(2)
const results = []
const chosen = SCENARIOS.filter((entry) => wanted.length === 0 || wanted.includes(entry.name))
for (const scenario of chosen) {
  results.push([scenario.name, await runScenario(scenario, executable, work)])
}
try { rmSync(work, { recursive: true, force: true }) } catch { /* the temp dir may still be locked by VS Code */ }
for (const [name, ok] of results) log(`${ok ? 'passed' : 'FAILED'}: ${name}`)
process.exit(results.every(([, ok]) => ok) ? 0 : 1)
