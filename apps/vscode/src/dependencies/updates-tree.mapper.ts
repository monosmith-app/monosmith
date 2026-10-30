import type { ListNode, ListResult } from '../list-view/index.ts'
import type { DependencyUpdate, UpdatesReport } from './updates-report.contract.ts'

const LEVELS: Array<{ level: string, label: string, color: string }> = [
  { level: 'patch', label: 'Patch', color: 'charts.green' },
  { level: 'minor', label: 'Minor', color: 'charts.yellow' },
  { level: 'major', label: 'Major', color: 'charts.red' },
]

/**
 * Lays the outdated dependencies out as smith's own report does: grouped
 * patch, minor and major, each with the version it moves to and where it is
 * declared, then the tools smith pins that have a newer release.
 */
export function toDependencyRows (report: UpdatesReport): ListResult {
  const updates = report.updates ?? []
  const tools = (report.tools ?? []).filter((tool) => tool.newer === true)
  const known = new Set(LEVELS.map((entry) => entry.level))
  const groups = [
    ...LEVELS,
    ...[...new Set(updates.map((update) => update.level))].filter((level) => !known.has(level)).map((level) => ({ level, label: level, color: 'charts.blue' })),
  ]
  const nodes: ListNode[] = [
    ...groups.flatMap((group): ListNode[] => {
      const inGroup = updates.filter((update) => update.level === group.level)
      if (inGroup.length === 0) return []

      return [{
        id:          `deps:${group.level}`,
        label:       group.label,
        description: String(inGroup.length),
        icon:        'arrow-up',
        color:       group.color,
        expanded:    group.level !== 'patch' || inGroup.length <= 8,
        children:    inGroup.map((update) => toUpdateRow(update, group.color)),
      }]
    }),
    ...(tools.length === 0
      ? []
      : [{
          id:          'deps:tools',
          label:       'Tools',
          description: String(tools.length),
          icon:        'tools',
          expanded:    true,
          children:    tools.map((tool) => ({
            id:          `deps:tool:${tool.plugin}:${tool.name}`,
            label:       tool.name,
            description: `→ ${tool.latest}`,
            icon:        'arrow-up',
            tooltip:     `${tool.name} is pinned at ${tool.pinned}; ${tool.latest} is out.`,
          })),
        }]),
  ]
  if (nodes.length === 0) return { nodes: [], message: 'Every dependency is up to date.', summary: 'up to date' }

  return { nodes, summary: `${updates.length} ${updates.length === 1 ? 'update' : 'updates'}` }
}

function toUpdateRow (update: DependencyUpdate, color: string): ListNode {
  const sites = update.sites ?? []
  const where = sites.map((site) => `${site.file} (${site.section}, ${site.spec})`)
  const rewritable = sites.length > 0 && sites.every((site) => site.rewritable !== false)

  return {
    id:          `deps:${update.plugin}:${update.name}`,
    label:       update.name,
    description: `→ ${update.latest}`,
    icon:        'arrow-up',
    color,
    ...(rewritable && { context: 'dependency', data: update.name }),
    tooltip:     [`${update.name} ${update.current} → ${update.latest} (${update.level})`, ...where].join('\n'),
  }
}
