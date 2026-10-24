import type { TaskStatus } from '../task-run/index.ts'
import type { WorkspaceSnapshot } from '../workspace/index.ts'

/** One row of the MonoSmith view. */
export interface TreeNode {
  id:          string
  kind:        'folder' | 'group' | 'project' | 'task'
  label:       string
  description: string
  /** The folder the row belongs to. */
  root:        string
  /** The projects a folder, group or project row stands for. */
  projects:    string[]
  project?:    string
  task?:       string
  status?:     TaskStatus
  affected?:   boolean
  /** What a screen reader says for the row. */
  spoken:      string
  /** What the pointer says for the row. */
  tooltip:     string
  children:    TreeNode[]
}

/** A folder of the editor, as the view shows it. */
export interface TreeFolder {
  root:      string
  name:      string
  /** What smith said; absent while the folder has no running smith. */
  snapshot?: WorkspaceSnapshot
  /** Why there is no snapshot, in a few words. */
  note?:     string
}

const GROUPS: Array<{ kind: string, label: string }> = [
  { kind: 'app', label: 'apps' },
  { kind: 'lib', label: 'libs' },
  { kind: 'package', label: 'packages' },
]

/**
 * Lays the workspace out as the view shows it: with one folder, a group per
 * project kind; with several, a row per folder first. Under a group the
 * projects by name, and under each project its tasks with the status the last
 * run gave them. With no running smith in any folder, nothing: the view then
 * shows its welcome content.
 * @param folders - the folders with a `smith.toml`
 * @param statusOf - the status of a task id (`api:build`) in a folder
 */
export function toTree (folders: readonly TreeFolder[], statusOf: (root: string, task: string) => TaskStatus | undefined): TreeNode[] {
  if (folders.every((folder) => folder.snapshot === undefined)) return []
  const first = folders[0]
  if (folders.length === 1 && first?.snapshot !== undefined) return toGroups(first.root, first.snapshot, statusOf)

  return folders.map((folder): TreeNode => {
    const children = folder.snapshot === undefined ? [] : toGroups(folder.root, folder.snapshot, statusOf)
    const projects = folder.snapshot?.projects.map((project) => project.name) ?? []
    const description = folder.snapshot === undefined ? (folder.note ?? '') : `${projects.length} projects`

    return {
      id:      `folder:${folder.root}`,
      kind:    'folder',
      label:   folder.name,
      description,
      root:    folder.root,
      projects,
      spoken:  `${folder.name}, ${description}`,
      tooltip: `${folder.root}${folder.note === undefined ? '' : `\n${folder.note}`}`,
      children,
    }
  })
}

function toGroups (root: string, snapshot: WorkspaceSnapshot, statusOf: (root: string, task: string) => TaskStatus | undefined): TreeNode[] {
  const known = new Set(GROUPS.map((group) => group.kind))
  const groups = [...GROUPS, ...[...new Set(snapshot.projects.map((p) => p.kind))].filter((kind) => !known.has(kind)).map((kind) => ({ kind, label: kind }))]

  return groups.flatMap((group) => {
    const projects = snapshot.projects.filter((project) => project.kind === group.kind).sort((a, b) => a.name.localeCompare(b.name))
    if (projects.length === 0) return []
    const children: TreeNode[] = projects.map((project) => {
      const affected = snapshot.affected.has(project.name)
      const description = [project.language, affected ? 'affected' : ''].filter(Boolean).join(' · ')

      return {
        id:       `project:${root}:${project.name}`,
        kind:     'project',
        label:    project.name,
        description,
        root,
        projects: [project.name],
        project:  project.name,
        affected,
        spoken:   `${project.name}, ${project.language === '' ? 'project' : project.language}, ${affected ? 'affected by the changes' : 'not affected'}, ${project.tasks.length} ${project.tasks.length === 1 ? 'task' : 'tasks'}`,
        tooltip:  `${project.name}\n${project.dir}${project.tasks.length > 0 ? '\nRun all its tasks from the context menu.' : ''}`,
        children: project.tasks.map((task) => {
          const status = statusOf(root, `${project.name}:${task}`)

          return {
            id:          `task:${root}:${project.name}:${task}`,
            kind:        'task',
            label:       task,
            description: status ?? '',
            root,
            projects:    [project.name],
            project:     project.name,
            task,
            status,
            spoken:      `${project.name} ${task}, ${status ?? 'not run yet'}. Press Enter to run.`,
            tooltip:     `${project.name}:${task}\nPress Enter to run.`,
            children:    [],
          }
        }),
      }
    })

    return {
      id:          `group:${root}:${group.kind}`,
      kind:        'group',
      label:       group.label,
      description: String(children.length),
      root,
      projects:    projects.map((project) => project.name),
      spoken:      `${group.label}, ${children.length} projects`,
      tooltip:     `${group.label}: ${children.length} projects`,
      children,
    } satisfies TreeNode
  })
}
