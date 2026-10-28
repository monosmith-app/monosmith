import type { TaskResult } from '../task-run/index.ts'
import type { WorkspaceProject, WorkspaceSnapshot } from '../workspace/index.ts'

/** One row of the MonoSmith view. */
export interface TreeNode {
  id:          string
  kind:        'folder' | 'group' | 'project' | 'task' | 'action'
  label:       string
  description: string
  /** A codicon id; the view tints it for a project. */
  icon:        string
  /** The folder the row belongs to. */
  root:        string
  /** The projects a folder, group or project row stands for. */
  projects:    string[]
  project?:    string
  task?:       string
  status?:     TaskResult['status']
  /** The changes affect the project. */
  affected?:   boolean
  /** Whether the row shows the affected dot: not when every project is affected, which tells nothing. */
  dot?:        boolean
  /** The task has a log to open. */
  hasLog?:     boolean
  /** Open the row, because the filter matched something under it. */
  reveal?:     boolean
  /** The command an `action` row runs when it is chosen. */
  command?:    string
  /** What a screen reader says for the row. */
  spoken:      string
  /** What the pointer says for the row. */
  tooltip:     string
  children:    TreeNode[]
}

/** A folder of the editor, as the view shows it. */
export interface TreeFolder {
  root:          string
  name:          string
  /** What smith said; absent while the folder has no running smith. */
  snapshot?:     WorkspaceSnapshot
  /** Why there is no snapshot, in a few words. */
  note?:         string
  /** Doctor says the folder lacks the MonoSmith agent files, and the person has not opted out. */
  agentMissing?: boolean
}

/** What the last run did to a task, and whether its output was kept. */
export type TaskFacts = TaskResult & { hasLog?: boolean }

/** How the person narrowed the view. */
export interface ViewFilter {
  /** Text a project or task name (or language) must contain; empty shows all. */
  text:         string
  /** Only the projects the changes affect. */
  affectedOnly: boolean
}

const NO_FILTER: ViewFilter = { text: '', affectedOnly: false }

const GROUPS: Array<{ kind: string, label: string, icon: string }> = [
  { kind: 'app', label: 'apps', icon: 'window' },
  { kind: 'lib', label: 'libs', icon: 'library' },
  { kind: 'package', label: 'packages', icon: 'package' },
]

const STATUS_ICONS: Record<TaskResult['status'], string> = {
  queued:  'circle-outline',
  running: 'sync~spin',
  passed:  'pass',
  failed:  'error',
  cached:  'history',
  skipped: 'debug-step-over',
}

/**
 * Lays the workspace out as the view shows it: with one folder, a group per
 * project kind; with several, a row per folder first. Under a group the
 * projects by name, and under each project its tasks with how the last run
 * ended. With no running smith in any folder, nothing: the view then shows its
 * welcome content.
 *
 * A project says its language and, when the changes affect it, shows a dot,
 * unless all of them are affected, when a dot on every row would tell nothing.
 * A task shows its icon and, beside it, only what the icon does not say: how
 * long it ran, or why it did not pass.
 * @param folders - the folders with a `smith.toml`
 * @param factsOf - what the last run did to a task id (`api:build`) in a folder
 * @param filter - how the person narrowed the view
 */
export function toTree (folders: readonly TreeFolder[], factsOf: (root: string, task: string) => TaskFacts | undefined, filter: ViewFilter = NO_FILTER): TreeNode[] {
  if (folders.every((folder) => folder.snapshot === undefined)) return []
  const first = folders[0]
  const narrowed = filter.text.trim() !== '' || filter.affectedOnly
  if (folders.length === 1 && first?.snapshot !== undefined) return [...agentRow(first, narrowed), ...toGroups(first.root, first.snapshot, factsOf, filter)]

  return folders.flatMap((folder): TreeNode[] => {
    const children = folder.snapshot === undefined ? [] : [...agentRow(folder, narrowed), ...toGroups(folder.root, folder.snapshot, factsOf, filter)]
    if (narrowed && children.length === 0) return []
    const projects = folder.snapshot?.projects.map((project) => project.name) ?? []
    const description = folder.snapshot === undefined ? (folder.note ?? '') : `${projects.length} projects`

    return [{
      id:      `folder:${folder.root}`,
      kind:    'folder',
      label:   folder.name,
      description,
      icon:    'root-folder',
      root:    folder.root,
      projects,
      spoken:  `${folder.name}, ${description}`,
      tooltip: `${folder.root}${folder.note === undefined ? '' : `\n${folder.note}`}`,
      children,
    }]
  })
}

/** The line that offers the agent in a folder whose doctor says it is missing; not while the view is narrowed. */
function agentRow (folder: TreeFolder, narrowed: boolean): TreeNode[] {
  if (narrowed || folder.agentMissing !== true) return []

  return [{
    id:          `agent:${folder.root}`,
    kind:        'action',
    label:       'Add the MonoSmith agent',
    description: 'with smith’s MCP server',
    icon:        'hubot',
    root:        folder.root,
    projects:    [],
    command:     'monosmith.installAgent',
    spoken:      'Add the MonoSmith agent and smith’s MCP server to this folder',
    tooltip:     'Writes .claude/agents/smith.md, .mcp.json and .vscode/mcp.json. It shows the files first and asks.',
    children:    [],
  }]
}

/**
 * Says how many projects the changes affect, for the title of the view:
 * `3 of 23 affected`, `all 23 affected`, or nothing when none are or it is not known.
 */
export function describeAffected (folders: readonly TreeFolder[]): string {
  const known = folders.flatMap((folder) => (folder.snapshot === undefined ? [] : [folder.snapshot]))
  const total = known.reduce((sum, snapshot) => sum + snapshot.projects.length, 0)
  const affected = known.reduce((sum, snapshot) => sum + snapshot.projects.filter((project) => snapshot.affected.has(project.name)).length, 0)
  if (total === 0 || affected === 0) return ''

  return affected === total ? `all ${total} affected` : `${affected} of ${total} affected`
}

/**
 * Says how the view is narrowed, or undefined when it is not. With nothing
 * shown it says why.
 * @param filter - how the person narrowed the view
 * @param shown - how many projects the view shows
 */
export function describeFilter (filter: ViewFilter, shown: number): string | undefined {
  const text = filter.text.trim()
  const parts = [filter.affectedOnly ? 'affected only' : '', text === '' ? '' : `“${text}”`].filter(Boolean)
  if (parts.length === 0) return undefined

  return shown === 0 ? `No project matches: ${parts.join(', ')}.` : `Showing ${parts.join(', ')}.`
}

/** How many project rows a tree holds. */
export function countProjects (nodes: readonly TreeNode[]): number {
  return nodes.reduce((sum, node) => sum + (node.kind === 'project' ? 1 : countProjects(node.children)), 0)
}

function toGroups (root: string, snapshot: WorkspaceSnapshot, factsOf: (root: string, task: string) => TaskFacts | undefined, filter: ViewFilter): TreeNode[] {
  const known = new Set(GROUPS.map((group) => group.kind))
  const groups = [
    ...GROUPS,
    ...[...new Set(snapshot.projects.map((p) => p.kind))].filter((kind) => !known.has(kind)).map((kind) => ({ kind, label: kind, icon: 'folder' })),
  ]
  const everyAffected = snapshot.projects.length > 0 && snapshot.projects.every((project) => snapshot.affected.has(project.name))
  const needle = filter.text.trim().toLowerCase()

  return groups.flatMap((group) => {
    const projects = snapshot.projects.filter((project) => project.kind === group.kind).sort((a, b) => a.name.localeCompare(b.name))
    const children = projects.flatMap((project): TreeNode[] => {
      const affected = snapshot.affected.has(project.name)
      if (!affected && filter.affectedOnly) return []
      const nameMatches = needle === '' || project.name.toLowerCase().includes(needle) || project.language.toLowerCase().includes(needle)
      const tasks = nameMatches ? project.tasks : project.tasks.filter((task) => task.toLowerCase().includes(needle))

      return !nameMatches && tasks.length === 0 ? [] : [toProject(root, project, tasks, { affected, dot: affected && !everyAffected, reveal: !nameMatches }, factsOf)]
    })
    if (children.length === 0) return []

    return {
      id:          `group:${root}:${group.kind}`,
      kind:        'group',
      label:       group.label,
      description: String(children.length),
      icon:        group.icon,
      root,
      projects:    children.flatMap((child) => child.projects),
      spoken:      `${group.label}, ${children.length} projects`,
      tooltip:     `${group.label}: ${children.length} projects`,
      children,
    } satisfies TreeNode
  })
}

function toProject (root: string, project: WorkspaceProject, tasks: readonly string[], shown: { affected: boolean, dot: boolean, reveal: boolean }, factsOf: (root: string, task: string) => TaskFacts | undefined): TreeNode {
  const language = project.language === '' ? 'project' : project.language
  const taskCount = `${project.tasks.length} ${project.tasks.length === 1 ? 'task' : 'tasks'}`

  return {
    id:          `project:${root}:${project.name}`,
    kind:        'project',
    label:       project.name,
    description: project.language,
    icon:        GROUPS.find((group) => group.kind === project.kind)?.icon ?? 'folder',
    root,
    projects:    [project.name],
    project:     project.name,
    affected:    shown.affected,
    dot:         shown.dot,
    reveal:      shown.reveal,
    spoken:      `${project.name}, ${language}, ${shown.affected ? 'affected by the changes' : 'not affected'}, ${taskCount}`,
    tooltip:     [
      project.name,
      `${project.kind}${project.language === '' ? '' : `, ${project.language}`}`,
      project.dir,
      shown.affected ? 'Affected by the changes.' : 'Not affected by the changes.',
      project.tasks.length > 0 ? 'Run all its tasks from the context menu.' : '',
    ].filter(Boolean).join('\n'),
    children: tasks.map((task) => toTask(root, project.name, task, factsOf(root, `${project.name}:${task}`))),
  }
}

function toTask (root: string, project: string, task: string, facts: TaskFacts | undefined): TreeNode {
  return {
    id:          `task:${root}:${project}:${task}`,
    kind:        'task',
    label:       task,
    description: describeResult(facts),
    icon:        facts === undefined ? 'play' : STATUS_ICONS[facts.status],
    root,
    projects:    [project],
    project,
    task,
    ...(facts !== undefined && { status: facts.status }),
    ...(facts?.hasLog === true && { hasLog: true }),
    spoken:      `${project} ${task}, ${spokenResult(facts)}. Press Enter to run.`,
    tooltip:     [`${project}:${task}`, ...tooltipResult(facts), 'Press Enter to run.'].join('\n'),
    children:    [],
  }
}

/** What the row says beside the icon: only what the icon does not. */
function describeResult (facts: TaskFacts | undefined): string {
  if (facts === undefined) return ''
  switch (facts.status) {
    case 'passed': {
      return (facts.durationMs ?? 0) > 0 ? formatDuration(facts.durationMs ?? 0) : ''
    }
    case 'failed':
    case 'skipped': {
      return facts.reason ?? ''
    }
    default: {
      return ''
    }
  }
}

function spokenResult (facts: TaskFacts | undefined): string {
  if (facts === undefined) return 'not run yet'
  const took = (facts.durationMs ?? 0) > 0 ? ` in ${formatDuration(facts.durationMs ?? 0)}` : ''

  return `${facts.status}${took}${facts.reason === undefined ? '' : `, ${facts.reason}`}`
}

function tooltipResult (facts: TaskFacts | undefined): string[] {
  if (facts === undefined) return ['Not run yet.']
  const took = (facts.durationMs ?? 0) > 0 ? ` in ${formatDuration(facts.durationMs ?? 0)}` : ''
  const headline: Record<TaskResult['status'], string> = {
    queued:  'Waiting to run.',
    running: 'Running.',
    passed:  `Passed${took}.`,
    failed:  `Failed${took}.`,
    cached:  'Cache hit: the result came from the cache, nothing ran.',
    skipped: 'Skipped.',
  }
  const when = facts.time === undefined || Number.isNaN(Date.parse(facts.time)) ? [] : [`Last run: ${new Date(facts.time).toLocaleString()}.`]

  return [headline[facts.status], ...(facts.reason === undefined ? [] : [`Reason: ${facts.reason}.`]), ...when, ...(facts.hasLog === true ? ['Its output is kept: use Show Output.'] : [])]
}

/** `850ms`, `1.2s`, `2m 05s`. */
export function formatDuration (ms: number): string {
  if (ms < 1000) return `${Math.round(ms)}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`
  const minutes = Math.floor(ms / 60_000)
  const seconds = Math.round((ms % 60_000) / 1000)

  return `${minutes}m ${String(seconds).padStart(2, '0')}s`
}
