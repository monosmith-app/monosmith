import type { ListNode, ListResult } from '../../shared/list-view/index.ts'
import type { DependencyUpdate, UpdatesReport } from './updates-report.contract.ts'
import { updateArguments } from './update-arguments.policy.ts'
import { say } from '../../shared/words/index.ts'

/** The mark of each level, as the terminal's Dependencies view colours it. */
const LEVELS: Record<string, { icon: string, color: string }> = {
  patch: { icon: 'circle-filled', color: 'charts.green' },
  minor: { icon: 'circle-filled', color: 'charts.blue' },
  major: { icon: 'warning', color: 'charts.yellow' },
}

/**
 * Lays the outdated dependencies out as the terminal's Dependencies view does: the patch and minor
 * releases, safe to take together, then the major ones, to take one at a time. A row says the
 * version it moves from and to, the level and where it is used; under it is the package's detail
 * (its versions, where it is declared, its registry, and the `smith deps up` command an update
 * runs). The tools smith pins that have a newer release come last.
 * @param report - what `updates` answered
 * @returns the rows, with the headline beside the title
 */
export function toDependencyRows (report: UpdatesReport): ListResult {
  const updates = report.updates ?? []
  const tools = (report.tools ?? []).filter((tool) => tool.newer === true)
  const safe = updates.filter((update) => update.level !== 'major')
  const major = updates.filter((update) => update.level === 'major')
  const nodes: ListNode[] = [
    ...(safe.length === 0 ? [] : [group('safe', say('depspage.group.safe', { count: safe.length }), 'pass', 'charts.green', safe)]),
    ...(major.length === 0 ? [] : [group('major', say('depspage.group.major', { count: major.length }), 'warning', 'charts.yellow', major)]),
    ...(tools.length === 0
      ? []
      : [{
          id:          'deps:tools',
          label:       say('vscode.dependencies.tools'),
          description: String(tools.length),
          icon:        'tools',
          expanded:    true,
          children:    tools.map((tool) => ({
            id:          `deps:tool:${tool.plugin}:${tool.name}`,
            label:       tool.name,
            description: `→ ${tool.latest}`,
            icon:        'arrow-up',
            tooltip:     say('vscode.dependencies.is-pinned-at-is-out', { name: tool.name, pinned: tool.pinned, latest: tool.latest }),
          })),
        }]),
  ]

  return nodes.length === 0 ? { nodes: [], message: say('vscode.dependencies.every-dependency-is-up-to-date'), summary: 'up to date' } : { nodes, summary: headline(updates) }
}

/** `↑ 3 with newer releases · 1 patch · 1 minor · 1 major`. */
function headline (updates: readonly DependencyUpdate[]): string {
  const levels = (['patch', 'minor', 'major'] as const)
    .map((level) => ({ level, count: updates.filter((update) => update.level === level).length }))
    .filter(({ count }) => count > 0)
    .map(({ level, count }) => say(`depspage.level.${level}`, { count }))

  return [say('depspage.headline', { count: updates.length }), ...levels].join(' · ')
}

function group (id: string, label: string, icon: string, color: string, updates: readonly DependencyUpdate[]): ListNode {
  return { id: `deps:${id}`, label, icon, color, expanded: true, children: updates.map((update) => toUpdateRow(update)) }
}

/**
 * Whether smith can rewrite every declaration of a dependency, so an update can be offered.
 * @param update - the dependency
 * @returns true when it has declarations and smith can rewrite all of them
 */
export function isRewritable (update: DependencyUpdate): boolean {
  const sites = update.sites ?? []

  return sites.length > 0 && sites.every((site) => site.rewritable !== false)
}

/** `apps/web +1`: the first file that declares the dependency, and how many more. */
function usedBy (update: DependencyUpdate): string {
  const files = [...new Set((update.sites ?? []).map((site) => site.file))]
  const [first] = files
  if (first === undefined) return ''

  return files.length === 1 ? first : say('depspage.used.more', { first, count: files.length - 1 })
}

function toUpdateRow (update: DependencyUpdate): ListNode {
  const sites = update.sites ?? []
  const where = sites.map((site) => `${site.file} (${site.section}, ${site.spec})`)
  const rewritable = isRewritable(update)
  const major = update.level === 'major'
  const mark = LEVELS[update.level] ?? { icon: 'arrow-up', color: 'charts.blue' }
  const context = major ? 'dependency.major' : 'dependency'
  const line = `smith ${(updateArguments(update.name, major) ?? ['deps', 'up', update.name]).join(' ')}`
  const id = `deps:${update.plugin}:${update.name}`
  const detail: ListNode[] = [
    { id: `${id}:versions`, label: `${update.current} → ${update.latest}`, description: update.level, icon: mark.icon, color: mark.color },
    ...(major ? [{ id: `${id}:major`, label: say('depspage.detail.major'), icon: 'warning', color: 'charts.yellow' }] : []),
    ...(sites.length === 0
      ? []
      : [{
          id:       `${id}:used`,
          label:    say('depspage.detail.used_by'),
          icon:     'references',
          expanded: true,
          children: sites.map((site, index) => ({ id: `${id}:used:${index}`, label: site.file, description: `${site.section} · ${site.spec}`, icon: 'file' })),
        },
        { id: `${id}:declared`, label: say('depspage.detail.declared'), description: [...new Set(sites.map((site) => site.section))].join(', '), icon: 'symbol-key' }]),
    ...(update.url === undefined || update.url === ''
      ? []
      : [{ id: `${id}:registry`, label: say('depspage.detail.registry'), description: update.url, icon: 'link-external', command: 'monosmith.dependencies.openRegistry', data: update.url, tooltip: update.url }]),
    { id: `${id}:same`, label: say('depspage.detail.same_as'), description: `$ ${line}`, icon: 'terminal', tooltip: line },
    ...(rewritable
      ? [{ id: `${id}:update`, label: say('depspage.button.update_one', { name: update.name }), description: `$ ${line}`, icon: 'arrow-up', context: `${context}.action`, data: update.name, command: 'monosmith.dependencies.update' }]
      : []),
  ]

  return {
    id,
    label:       update.name,
    description: [`${update.current} → ${update.latest}`, update.level, usedBy(update)].filter(Boolean).join(' · '),
    icon:        mark.icon,
    color:       mark.color,
    ...(rewritable && { context, data: update.name }),
    tooltip:     [`${update.name} ${update.current} → ${update.latest} (${update.level})`, ...where, ...(update.url === undefined || update.url === '' ? [] : [`Registry: ${update.url}`]), `$ ${line}`].join('\n'),
    children:    detail,
  }
}
