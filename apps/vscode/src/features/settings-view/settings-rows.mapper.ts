import type { ListNode, ListResult } from '../../shared/list-view/index.ts'
import { commandLineOf } from '../command-palette/index.ts'
import type { PluginEntry, RecipeEntry, SettingsSections, WorkspaceSetting } from './settings-sections.contract.ts'
import { say, sayCount, type WordKey } from '../../shared/words/index.ts'

/** The type of a setting in the words of the terminal's table. */
function typeWord (setting: WorkspaceSetting): string {
  if (setting.type === 'boolean') return say('settingspage.type.switch')
  if ((setting.options ?? []).length > 0) return say(setting.type === 'array' ? 'settingspage.type.ticks' : 'settingspage.type.choice')
  if (setting.type === 'array') return say('settingspage.type.list')
  if (setting.type === 'integer' || setting.type === 'number') return say('settingspage.type.number')

  return say('settingspage.type.text')
}

/** The words of a value as smith.toml writes it: `["github", "azure"]` is two, `"warn"` or `warn` one. */
function words (raw: string): string[] {
  const quoted = Array.from(raw.matchAll(/"([^"]*)"/g), (match) => match[1] ?? '')
  if (quoted.length > 0) return quoted

  return raw.trim() === '' ? [] : [raw.trim()]
}

/** The default of a setting when smith knows it: the options marked as the default. */
function defaultOf (setting: WorkspaceSetting): string | undefined {
  const marked = (setting.options ?? []).filter((option) => option.default === true).map((option) => option.value)

  return marked.length === 0 ? undefined : marked.join(', ')
}

/**
 * What the Value column of the terminal shows for a setting: `● on` or `○ off` for a switch,
 * the value with `▾` for a fixed set, `not decided` for a switch with no value and no default,
 * and the default for any other setting that is not set.
 * @param setting - the setting
 * @returns the value as the row shows it
 */
export function valueText (setting: WorkspaceSetting): string {
  const value = setting.set ? words(setting.value ?? '').join(', ') : undefined
  if (setting.type === 'boolean') {
    if (value === undefined) return `○ ${say('settingspage.value.undecided')}`

    return value === 'true' ? `● ${say('vscode.settings-view.on')}` : `○ ${say('vscode.settings-view.off')}`
  }
  const shown = value ?? say('vscode.settings-view.default-value', { value: defaultOf(setting) ?? say('vscode.settings-view.none') })

  return (setting.options ?? []).length > 0 && setting.type !== 'array' ? `${shown} ▾` : shown
}

/**
 * The command line a change of the setting stands for, as the terminal's Settings box shows it.
 * @param setting - the setting
 * @returns `smith config set <path> <value>`, with the value it holds now, or `<value>`
 */
export function setCommandLine (setting: WorkspaceSetting): string {
  const value = setting.set ? words(setting.value ?? '').join(' ') : ''

  return `smith config set ${setting.path} ${value === '' ? '<value>' : value}`
}

function settingRow (setting: WorkspaceSetting): ListNode {
  const line = setCommandLine(setting)
  const fallback = defaultOf(setting)

  return {
    id:          `settings:setting:${setting.path}`,
    label:       setting.path,
    description: `${valueText(setting)} · ${typeWord(setting)}`,
    icon:        setting.set ? 'circle-filled' : 'circle-outline',
    ...(setting.set && { color: 'charts.blue' }),
    context:     setting.set ? 'setting.set' : 'setting',
    data:        setting.path,
    command:     'monosmith.settings.change',
    tooltip:     [
      setting.description,
      setting.reason === undefined ? '' : `${say('vscode.settings-view.why-default')} ${setting.reason}`,
      fallback === undefined ? '' : say('vscode.settings-view.default-value', { value: fallback }),
      `$ ${line}`,
      say('settingspage.writes'),
    ].filter(Boolean).join('\n'),
  }
}

function pluginRow (plugin: PluginEntry): ListNode {
  return {
    id:          `settings:plugin:${plugin.name}`,
    label:       plugin.name,
    description: [plugin.source, (plugin.roles ?? []).join(', ')].filter(Boolean).join(' · '),
    icon:        'extensions',
    context:     'setting.plugin',
    data:        plugin.name,
    command:     'monosmith.settings.plugin',
    tooltip:     `$ smith plugin enable ${plugin.name}\n$ smith plugin disable ${plugin.name}`,
  }
}

function recipeRow (recipe: RecipeEntry): ListNode {
  return {
    id:          `settings:recipe:${recipe.path}`,
    label:       recipe.name ?? recipe.path,
    description: recipe.error ?? [recipe.version, recipe.path].filter(Boolean).join(' · '),
    icon:        recipe.error === undefined ? 'file-code' : 'warning',
    ...(recipe.error !== undefined && { color: 'charts.yellow' }),
    data:        recipe.path,
    command:     'monosmith.settings.openRecipe',
    tooltip:     recipe.error ?? `$ smith recipe check ${recipe.path}`,
  }
}

/** A row that runs a command of the extension, with its smith command line beside it. */
function commandRow (label: WordKey, command: string, icon: string): ListNode {
  const line = commandLineOf(command)

  return { id: `settings:command:${command}`, label: say(label), ...(line !== undefined && { description: `$ ${line}`, tooltip: line }), icon, command }
}

function section (id: string, label: WordKey, description: string, icon: string, children: ListNode[], expanded = false): ListNode {
  return { id: `settings:${id}`, label: say(label), description, icon, ...(expanded && { expanded }), children }
}

/**
 * Lays the settings out as the terminal's Settings view does, as sections: Plugins, Architecture,
 * Workspace (every setting of smith.toml with its value, type and the `smith config set` command
 * it stands for), Recipes, Assistants, CI and Forms. A change is never made from the row: it
 * opens the question and the confirmation, which show the command first.
 * @param sections - what smith answered
 * @returns the rows, with the count of settings beside the title
 */
export function toSettingsRows (sections: SettingsSections): ListResult {
  const set = sections.settings.filter((setting) => setting.set).length
  const nodes: ListNode[] = [
    section('plugins', 'settingsview.section.plugins', sayCount('vscode.settings-view.plugins', sections.plugins.length), 'extensions', [
      ...sections.plugins.map((plugin) => pluginRow(plugin)),
      commandRow('vscode.settings-view.install-plugin', 'monosmith.pluginInstall', 'cloud-download'),
    ]),
    section('architecture', 'settingsview.section.architecture', '', 'layers', [commandRow('vscode.settings-view.choose-architecture', 'monosmith.architecture', 'layers')]),
    section('workspace', 'settingsview.section.workspace', sayCount('settingspage.sub.workspace', sections.settings.length, { set }), 'settings-gear', sections.settings.map((setting) => settingRow(setting)), true),
    section('recipes', 'settingsview.section.recipes', sayCount('settingspage.sub.recipes', sections.recipes.length), 'book', [
      ...sections.recipes.map((recipe) => recipeRow(recipe)),
      commandRow('vscode.settings-view.new-recipe', 'monosmith.newRecipe', 'add'),
    ]),
    section('assistants', 'settingsview.section.assistants', '', 'hubot', [
      commandRow('vscode.settings-view.install-agent', 'monosmith.installAgent', 'hubot'),
      commandRow('vscode.settings-view.connect-assistant', 'monosmith.mcpInstall', 'plug'),
    ]),
    section('ci', 'settingsview.section.ci', '', 'rocket', [
      commandRow('vscode.settings-view.regenerate-ci', 'monosmith.ciGenerate', 'sync'),
      commandRow('vscode.settings-view.add-pipeline', 'monosmith.pipeline', 'add'),
    ]),
    section('forms', 'settingsview.section.forms', '', 'list-flat', [
      commandRow('dashboard.add.label', 'monosmith.addProject', 'add'),
      commandRow('dashboard.compose.label', 'monosmith.composeStack', 'layers'),
      commandRow('vscode.settings-view.remove-project', 'monosmith.removeProject', 'trash'),
      commandRow('vscode.settings-view.change-kind', 'monosmith.changeKind', 'symbol-class'),
      commandRow('vscode.settings-view.swap-choice', 'monosmith.swap', 'arrow-swap'),
      commandRow('vscode.settings-view.sync-versions', 'monosmith.syncVersions', 'sync'),
    ]),
    { id: 'settings:hint', label: say('settingspage.sections.hint'), icon: 'info', command: 'monosmith.projects.focus' },
  ]

  return { nodes, ...(sections.errors.length > 0 && { message: sections.errors.join(' ') }), summary: sayCount('settingspage.sub.workspace', sections.settings.length, { set }) }
}
