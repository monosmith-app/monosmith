import { describe, expect, it } from 'vitest'
import { setCommandLine, toSettingsRows, valueText } from './settings-rows.mapper.ts'

const settings = [
  { path: 'workspace.commit_operations', type: 'boolean', description: 'Commit every mutating command.', reason: 'smith never commits unless you ask.', set: false },
  { path: 'workspace.confirm', type: 'string', description: 'Which plans ask.', set: true, value: '"risky"', options: [{ value: 'always', default: true }, { value: 'risky' }] },
  { path: 'ci.providers', type: 'array', description: 'The CI providers.', set: true, value: '["github", "azure"]', options: [{ value: 'github' }, { value: 'azure' }] },
  { path: 'release.tag_prefix', type: 'string', description: 'The tag prefix.', set: false },
]
const sections = {
  settings,
  plugins: [{ name: 'go', source: 'built-in', roles: ['language', 'toolchain'] }],
  recipes: [{ path: 'team.recipe.toml', name: 'team', version: '1.3.0' }, { path: 'bad.recipe.toml', error: 'not a recipe' }],
  errors:  [],
}

describe('valueText', () => {
  it('shows a switch as on or off, and not decided while it is not set', () => {
    expect(valueText({ ...settings[0], set: true, value: 'true' })).toBe('● on')
    expect(valueText({ ...settings[0], set: true, value: 'false' })).toBe('○ off')
    expect(valueText(settings[0])).toBe('○ not decided')
  })

  it('shows a value of a fixed set with its marker, a list as its words, and the default of what is not set', () => {
    expect(valueText(settings[1])).toBe('risky ▾')
    expect(valueText(settings[2])).toBe('github, azure')
    expect(valueText({ ...settings[1], set: false })).toBe('default: always ▾')
    expect(valueText(settings[3])).toBe('default: none')
  })
})

describe('setCommandLine', () => {
  it('is the smith config set command the setting stands for', () => {
    expect(setCommandLine(settings[2])).toBe('smith config set ci.providers github azure')
    expect(setCommandLine(settings[3])).toBe('smith config set release.tag_prefix <value>')
  })
})

describe('toSettingsRows', () => {
  it('has the sections of the terminal, Workspace open, and the hint about project settings last', () => {
    const { nodes, summary } = toSettingsRows(sections)
    expect(nodes.map((node) => node.label)).toEqual(['Plugins', 'Architecture', 'Workspace', 'Recipes', 'Assistants', 'CI', 'Forms', 'Per-project settings are in Projects › a project › Settings'])
    expect(nodes[2]).toMatchObject({ description: '4 settings, 2 set', expanded: true })
    expect(summary).toBe('4 settings, 2 set')
  })

  it('gives each setting its value, its type, and its command, and a reset only when it is set', () => {
    const rows = toSettingsRows(sections).nodes[2]?.children ?? []
    expect(rows[0]).toMatchObject({ label: 'workspace.commit_operations', description: '○ not decided · switch', context: 'setting', data: 'workspace.commit_operations', command: 'monosmith.settings.change' })
    expect(rows[0]?.tooltip).toContain('$ smith config set workspace.commit_operations <value>')
    expect(rows[0]?.tooltip).toContain('Writes smith.toml exactly as the command does.')
    expect(rows[1]).toMatchObject({ description: 'risky ▾ · choice', context: 'setting.set' })
    expect(rows[2]?.description).toBe('github, azure · ticks')
  })

  it('lists the plugins with a switch, and the recipes with what is wrong with one', () => {
    const { nodes } = toSettingsRows(sections)
    const plugins = nodes[0]
    const recipes = nodes[3]
    expect(plugins?.children?.[0]).toMatchObject({ label: 'go', description: 'built-in · language, toolchain', context: 'setting.plugin', command: 'monosmith.settings.plugin' })
    expect(plugins?.children?.at(-1)).toMatchObject({ command: 'monosmith.pluginInstall', description: '$ smith plugin install <plugin>' })
    expect(recipes?.children?.slice(0, 2)).toMatchObject([
      { label: 'team', description: '1.3.0 · team.recipe.toml', command: 'monosmith.settings.openRecipe', data: 'team.recipe.toml' },
      { label: 'bad.recipe.toml', description: 'not a recipe', icon: 'warning' },
    ])
  })

  it('says what could not be read', () => {
    expect(toSettingsRows({ ...sections, errors: ['method not found'] }).message).toBe('method not found')
  })
})
