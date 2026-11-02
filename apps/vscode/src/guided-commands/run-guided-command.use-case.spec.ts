import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import type { ResolvedStep } from './guided-command.contract.ts'
import { GUIDED_COMMANDS } from './guided-commands.config.ts'
import { parsePipelineTemplates, parseWebApps, parseWorkspaceEnvironments } from './guided-sources.mapper.ts'
import { runGuidedCommand, type GuidedDeps } from './run-guided-command.use-case.ts'

const ok = (stdout: string): SmithRun => ({ code: 0, stdout, stderr: '' })
const APPS = JSON.stringify([
  { project: 'web', envFiles: 'commit', output: 'workspace', default: 'prd', environments: [{ name: 'dev', task: 'build:dev' }, { name: 'prd', task: 'build' }] },
])

interface Fake { deps: GuidedDeps, calls: string[][], asked: ResolvedStep[], shown: string[], confirmed: string[] }

/** A fake editor: `answers` answer the steps in order, `yes` answers the confirmation, `runs` answers smith by its first words. */
function fake (answers: Array<string | string[] | undefined>, yes = true, runs: Record<string, SmithRun> = {}): Fake {
  const f: Fake = { deps: undefined as never, calls: [], asked: [], shown: [], confirmed: [] }
  const queue = [...answers]
  f.deps = {
    run: async (args) => {
      f.calls.push(args)

      return runs[args.slice(0, 2).join(' ')] ?? ok('')
    },
    ask: async (step) => {
      f.asked.push(step)

      return queue.shift()
    },
    show:    async (_title, text) => { f.shown.push(text) },
    confirm: async (_title, detail) => {
      f.confirmed.push(detail)

      return yes
    },
  }

  return f
}

const command = (id: string): (typeof GUIDED_COMMANDS)[number] => GUIDED_COMMANDS.find((c) => c.id === id) ?? GUIDED_COMMANDS[0]

describe('adding environments to a web app', () => {
  it('lists the web apps from smith, starts the names on the workspace\'s, shows the plan and runs it with --yes only on a yes', async () => {
    const f = fake(['web', 'qa, stage'], true, {
      'environments list': ok(APPS),
      'config get':        ok('["lcl", "dev"]\n'),
      'environments add':  ok('Wrote 3 file(s):\n  web/.env.qa\n'),
    })
    const outcome = await runGuidedCommand(command('environmentsAdd'), f.deps)
    expect(f.asked[0].options).toEqual([{ label: 'web', value: 'web', description: 'dev, prd' }])
    expect(f.asked[1].value).toBe('lcl, dev')
    expect(f.calls.at(-2)).toEqual(['environments', 'add', 'web', 'qa', 'stage', '--dry-run'])
    expect(f.calls.at(-1)).toEqual(['environments', 'add', 'web', 'qa', 'stage', '--yes'])
    expect(f.confirmed[0]).toContain('Nothing was changed yet.')
    expect(outcome).toMatchObject({ kind: 'done' })
  })

  it('starts the names on the default four when the workspace names none', async () => {
    const f = fake(['web', undefined], true, { 'environments list': ok(APPS), 'config get': { code: 1, stdout: '', stderr: 'not set' } })
    await runGuidedCommand(command('environmentsAdd'), f.deps)
    expect(f.asked.at(-1)?.value).toBe('lcl, dev, uat, prd')
  })

  it('changes nothing when the person says no', async () => {
    const f = fake(['web', 'qa'], false, { 'environments list': ok(APPS) })
    expect(await runGuidedCommand(command('environmentsAdd'), f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.calls.some((args) => args.includes('--yes'))).toBe(false)
  })

  it('says how to make a web app when there is none', async () => {
    const f = fake([], true, { 'environments list': ok('[]') })
    expect(await runGuidedCommand(command('environmentsAdd'), f.deps)).toMatchObject({ kind: 'refused', message: expect.stringContaining('--environments') })
  })
})

describe('removing environments', () => {
  it('offers the environments of the app except the default one, which cannot be removed', async () => {
    const f = fake(['web', ['dev']], true, { 'environments list': ok(APPS) })
    await runGuidedCommand(command('environmentsRemove'), f.deps)
    expect(f.asked[1].options.map((o) => o.value)).toEqual(['dev'])
    expect(f.calls.at(-1)).toEqual(['environments', 'remove', 'web', 'dev', '--yes'])
  })
})

describe('the other commands', () => {
  it('shows the output of a read command and changes nothing', async () => {
    const f = fake([], true, { 'cache status': ok('local cache: 12 MB') })
    expect(await runGuidedCommand(command('cacheStatus'), f.deps)).toEqual({ kind: 'shown' })
    expect(f.shown).toEqual(['local cache: 12 MB'])
  })

  it('shows the consequence of a change with no plan, and asks first', async () => {
    const f = fake([], true)
    await runGuidedCommand(command('cacheClear'), f.deps)
    expect(f.confirmed[0]).toContain('The remote cache stays')
    expect(f.calls).toEqual([['cache', 'clear', '--yes']])
  })

  it('returns the words of smith when it refuses', async () => {
    const f = fake([], true, { 'cache status': { code: 1, stdout: '', stderr: '✖ no cache here\n' } })
    expect(await runGuidedCommand(command('cacheStatus'), f.deps)).toEqual({ kind: 'refused', message: 'no cache here' })
  })

  it('builds the commands of the forms from the answers', async () => {
    const pipeline = fake(['deploy-pages', 'web'], true, { pipeline: ok('deploy-pages (--project <app>)  github  build an app\n') })
    await runGuidedCommand(command('pipeline'), pipeline.deps)
    expect(pipeline.calls.at(-1)).toEqual(['pipeline', 'deploy-pages', '--project', 'web', '--yes'])
    const icons = fake([String.raw`C:\logo.png`, ''])
    await runGuidedCommand(command('tauriIcon'), icons.deps)
    expect(icons.calls.at(-1)).toEqual(['tauri', 'icon', String.raw`C:\logo.png`, '--yes'])
    const why = fake(['web:build'], true, { 'why-miss web:build': ok('inputs changed') })
    await runGuidedCommand(command('whyMiss'), why.deps)
    expect(why.calls).toEqual([['why-miss', 'web:build']])
    const mcp = fake([['--claude', '--vscode']])
    await runGuidedCommand(command('mcpInstall'), mcp.deps)
    expect(mcp.calls).toEqual([['mcp', 'install', '--claude', '--vscode', '--yes']])
  })
})

const SETTINGS = JSON.stringify([
  { path: 'ci.devcontainer', type: 'boolean', description: 'Write a dev container.', reason: 'it makes a laptop and CI agree', value: 'true', set: true },
  { path: 'release.api_check', type: 'string', description: 'How the public API is checked.', enum: ['warn', 'fail', 'off'], value: '', set: false },
  { path: 'workspace.scope', type: 'string', description: 'The npm scope.', value: '', set: false },
])

describe('the settings of smith.toml', () => {
  it('picks a setting, then a boolean from true and false with the current one marked, and writes it without --yes', async () => {
    const f = fake(['ci.devcontainer', 'false'], true, { 'config list': ok(SETTINGS), 'config set': ok('✔ ci.devcontainer = false\n') })
    expect(await runGuidedCommand(command('configSet'), f.deps)).toMatchObject({ kind: 'done' })
    expect(f.asked[1].kind).toBe('pick')
    expect(f.asked[1].options.map((o) => o.label)).toEqual(['true (current)', 'false'])
    expect(f.calls.at(-1)).toEqual(['config', 'set', 'ci.devcontainer', 'false'])
  })

  it('offers the values of an enumerated setting, and a text box with the current value for any other', async () => {
    const enumerated = fake(['release.api_check', 'warn'], true, { 'config list': ok(SETTINGS) })
    await runGuidedCommand(command('configSet'), enumerated.deps)
    expect(enumerated.asked[1].options.map((o) => o.value)).toEqual(['warn', 'fail', 'off'])
    const typed = fake(['workspace.scope', '@acme'], true, { 'config list': ok(SETTINGS) })
    await runGuidedCommand(command('configSet'), typed.deps)
    expect(typed.asked[1]).toMatchObject({ kind: 'text', value: '' })
  })

  it('resets only a setting that smith.toml writes', async () => {
    const f = fake(['ci.devcontainer'], true, { 'config list': ok(SETTINGS) })
    await runGuidedCommand(command('configUnset'), f.deps)
    expect(f.asked[0].options.map((o) => o.value)).toEqual(['ci.devcontainer'])
    expect(f.calls.at(-1)).toEqual(['config', 'unset', 'ci.devcontainer'])
  })
})

describe('plugins and the architecture', () => {
  it('turns a built-in plugin off, and carries the warning smith prints about the projects it leaves', async () => {
    const plugins = JSON.stringify([{ name: 'node', source: 'built-in', provides: ['node'] }, { name: 'ext', source: '/x' }])
    const f = fake(['node'], true, {
      'plugin list':    ok(plugins),
      'plugin disable': { code: 0, stdout: 'enabled plugins: python\n', stderr: 'warning: 3 projects disappear: web, api, sdk\n' },
    })
    const outcome = await runGuidedCommand(command('pluginDisable'), f.deps)
    expect(f.asked[0].options.map((o) => o.value)).toEqual(['node'])
    expect(f.confirmed[0]).toContain('disappear')
    expect(outcome).toMatchObject({ kind: 'done', lines: ['enabled plugins: python', 'warning: 3 projects disappear: web, api, sdk'] })
  })

  it('lists the architectures with the one in use marked', async () => {
    const f = fake(['none'], true, { architecture: ok('* vertical-slices  one folder per feature\n                   applies to: node\n  none             no checks\n') })
    await runGuidedCommand(command('architecture'), f.deps)
    expect(f.asked[0].options).toEqual([
      { label: 'vertical-slices (in use)', value: 'vertical-slices', description: 'one folder per feature' },
      { label: 'none', value: 'none', description: 'no checks' },
    ])
    expect(f.calls.at(-1)).toEqual(['architecture', 'none'])
  })

  it('regenerates the pipelines after showing the plan', async () => {
    const f = fake([], true, { 'ci generate': ok('plan') })
    await runGuidedCommand(command('ciGenerate'), f.deps)
    expect(f.calls).toEqual([['ci', 'generate', '--dry-run'], ['ci', 'generate', '--yes']])
  })
})

describe('installing a plugin from the marketplace', () => {
  const found = JSON.stringify([{ name: 'deployer', newest: '0.2.0', description: 'Deploys to the moon', roles: ['deploy'] }])

  it('lists what the search found, asks the trust question and installs, pinned only on that answer', async () => {
    const f = fake(['deploy', 'deployer', 'pinned'], true, { 'plugin search': ok(found), 'plugin install': ok('installed deployer\n') })
    await runGuidedCommand(command('pluginInstall'), f.deps)
    expect(f.calls[0]).toEqual(['plugin', 'search', 'deploy', '--json'])
    expect(f.asked[1].options).toEqual([{ label: 'deployer', value: 'deployer', description: 'v0.2.0 · Deploys to the moon [deploy]' }])
    expect(f.confirmed[0]).toContain('does not sandbox it')
    expect(f.calls.at(-1)).toEqual(['plugin', 'install', 'deployer', '--trust', '--pin'])
    const plain = fake(['', 'deployer', 'ask'], true, { 'plugin search': ok(found) })
    await runGuidedCommand(command('pluginInstall'), plain.deps)
    expect(plain.calls.at(-1)).toEqual(['plugin', 'install', 'deployer'])
  })

  it('says so when the marketplace is empty or nothing matches', async () => {
    const empty = fake([''], true, { 'plugin search': ok('[]') })
    expect(await runGuidedCommand(command('pluginInstall'), empty.deps)).toMatchObject({ kind: 'refused', message: expect.stringContaining('lists no plugins yet') })
    const none = fake(['zzz'], true, { 'plugin search': ok('[]') })
    expect(await runGuidedCommand(command('pluginInstall'), none.deps)).toMatchObject({ kind: 'refused', message: expect.stringContaining('No plugin') })
  })
})

describe('reading what smith prints', () => {
  it('reads web apps, workspace environments and pipelines, and falls back when the text is not what it expects', () => {
    expect(parseWebApps(APPS)[0].environments.map((e) => e.name)).toEqual(['dev', 'prd'])
    expect(parseWebApps('nope')).toEqual([])
    expect(parseWorkspaceEnvironments('["qa", "prd"]')).toEqual(['qa', 'prd'])
    expect(parseWorkspaceEnvironments('')).toEqual(['lcl', 'dev', 'uat', 'prd'])
    expect(parsePipelineTemplates('e2e  github, azure  run every e2e task\ndeploy-pages (--project <app>)  github  deploy\n')).toEqual([
      { label: 'e2e', value: 'e2e', description: 'run every e2e task [github, azure]' },
      { label: 'deploy-pages', value: 'deploy-pages', description: 'deploy (asks for the app) [github]' },
    ])
  })
})
