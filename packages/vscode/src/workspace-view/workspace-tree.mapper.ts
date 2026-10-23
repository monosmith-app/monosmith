import type { TaskStatus } from '../task-run/index.ts'
import type { WorkspaceSnapshot } from '../workspace/index.ts'

/** One row of the MonoSmith view. */
export interface TreeNode {
  id:          string
  kind:        'group' | 'project' | 'task'
  label:       string
  description: string
  /** The project a project or task row belongs to. */
  project?:    string
  task?:       string
  status?:     TaskStatus
  affected?:   boolean
  children:    TreeNode[]
}

const GROUPS: Array<{ kind: string, label: string }> = [
  { kind: 'app', label: 'apps' },
  { kind: 'lib', label: 'libs' },
  { kind: 'package', label: 'packages' },
]

/**
 * Lays the workspace out as the view shows it: a group per project kind, the
 * projects in it by name, and under each project its tasks with the status the
 * last run gave them.
 * @param snapshot - the projects and what is affected
 * @param statusOf - the status of a task id (`api:build`)
 */
export function toTree (snapshot: WorkspaceSnapshot, statusOf: (task: string) => TaskStatus | undefined): TreeNode[] {
  const known = new Set(GROUPS.map((group) => group.kind))
  const groups = [...GROUPS, ...[...new Set(snapshot.projects.map((p) => p.kind))].filter((kind) => !known.has(kind)).map((kind) => ({ kind, label: kind }))]

  return groups.flatMap((group) => {
    const projects = snapshot.projects.filter((project) => project.kind === group.kind).sort((a, b) => a.name.localeCompare(b.name))
    if (projects.length === 0) return []
    const children: TreeNode[] = projects.map((project) => {
      const affected = snapshot.affected.has(project.name)

      return {
        id:          `project:${project.name}`,
        kind:        'project',
        label:       project.name,
        description: [project.language, affected ? 'affected' : ''].filter(Boolean).join(' · '),
        project:     project.name,
        affected,
        children:    project.tasks.map((task) => {
          const status = statusOf(`${project.name}:${task}`)

          return {
            id:          `task:${project.name}:${task}`,
            kind:        'task',
            label:       task,
            description: status ?? '',
            project:     project.name,
            task,
            status,
            children:    [],
          }
        }),
      }
    })

    return { id: `group:${group.kind}`, kind: 'group', label: group.label, description: String(children.length), children }
  })
}
