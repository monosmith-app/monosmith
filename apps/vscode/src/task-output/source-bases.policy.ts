import { join } from 'node:path'

/** A folder of the editor and where its projects are. */
export interface FolderProjects {
  root:     string
  projects: ReadonlyArray<{ name: string, dir: string }>
}

/**
 * The directories a relative path printed by a task may be relative to, the
 * likeliest first: the directory of the project the task belongs to, then the
 * folder it is in. A task id is `<project>:<task>`; without one (a line that
 * says no task) only the folders are tried.
 * @param folders - the folders that run smith
 * @param taskId - the task the line belongs to, if it says
 * @param root - the folder the task ran in, if known; else every folder is tried
 */
export function sourceBases (folders: readonly FolderProjects[], taskId: string | undefined, root?: string): string[] {
  const project = taskId?.split(':', 1)[0]
  const candidates = root === undefined ? folders : folders.filter((folder) => folder.root === root)
  const dirs = project === undefined
    ? []
    : candidates.flatMap((folder) => folder.projects.filter((entry) => entry.name === project).map((entry) => join(folder.root, entry.dir)))

  return [...dirs, ...candidates.map((folder) => folder.root)]
}
