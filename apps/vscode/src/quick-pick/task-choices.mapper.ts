import type { FolderWorkspace, WorkspaceProject, WorkspaceSnapshot } from '../workspace/index.ts'

/** One line of a quick pick, carrying what it stands for. */
export interface Choice<T> {
  label:        string
  description?: string
  detail?:      string
  value:        T
}

/** The folders to choose from, when there are several. */
export function folderChoices (folders: readonly FolderWorkspace[]): Array<Choice<string>> {
  return folders.map((folder) => ({
    label:       folder.name,
    description: `${folder.snapshot.projects.length} projects`,
    value:       folder.root,
  }))
}

/** The projects of a folder: those the changes affect first, then by name. */
export function projectChoices (snapshot: WorkspaceSnapshot): Array<Choice<WorkspaceProject>> {
  return [...snapshot.projects]
    .filter((project) => project.tasks.length > 0)
    .sort((a, b) => Number(snapshot.affected.has(b.name)) - Number(snapshot.affected.has(a.name)) || a.name.localeCompare(b.name))
    .map((project) => ({
      label:       snapshot.affected.has(project.name) ? `$(circle-filled) ${project.name}` : project.name,
      description: [project.language, snapshot.affected.has(project.name) ? 'affected' : ''].filter(Boolean).join(' · '),
      detail:      project.tasks.join(', '),
      value:       project,
    }))
}

/** What to run in a project: all its tasks, or one of them. `undefined` stands for all. */
export function taskChoices (project: WorkspaceProject): Array<Choice<string | undefined>> {
  return [
    { label: '$(run-all) All tasks', description: project.tasks.join(', '), value: undefined },
    ...project.tasks.map((task) => ({ label: task, value: task })),
  ]
}

/**
 * The task names to run across projects, each with the number of projects
 * that have it. With `onlyAffected`, only the affected projects count.
 */
export function taskNameChoices (snapshot: WorkspaceSnapshot, onlyAffected: boolean): Array<Choice<string>> {
  const counts = new Map<string, number>()
  for (const project of snapshot.projects) {
    if (onlyAffected && !snapshot.affected.has(project.name)) continue
    for (const task of project.tasks) counts.set(task, (counts.get(task) ?? 0) + 1)
  }

  return [...counts]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([task, count]) => ({ label: task, description: `${count} ${count === 1 ? 'project' : 'projects'}`, value: task }))
}
