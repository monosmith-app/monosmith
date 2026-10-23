import type { SmithConnection } from '../smith-client/index.ts'
import type { WorkspaceProject, WorkspaceSnapshot } from './workspace-snapshot.contract.ts'

interface ProjectsAnswer {
  name:      string
  dir?:      string
  root?:     string
  kind?:     string
  language?: string
  tasks?:    string[]
  targets?:  string[]
  deps?:     string[]
}

/**
 * Asks smith for the projects and for what is affected by the changes.
 * Affected is optional knowledge: when it cannot be had (no git history, say)
 * the projects still show.
 */
export async function readWorkspace (connection: SmithConnection): Promise<WorkspaceSnapshot> {
  const [projects, affected] = await Promise.all([
    connection.request<ProjectsAnswer[]>('projects'),
    affectedProjects(connection),
  ])

  return {
    projects: projects.map((project) => toProject(project)),
    affected: new Set(affected),
  }
}

async function affectedProjects (connection: SmithConnection): Promise<string[]> {
  try {
    const answer = await connection.request<{ projects?: string[] }>('affected')

    return answer.projects ?? []
  } catch {
    return []
  }
}

function toProject (answer: ProjectsAnswer): WorkspaceProject {
  return {
    name:     answer.name,
    dir:      answer.dir ?? answer.root ?? '',
    kind:     answer.kind ?? 'app',
    language: answer.language ?? '',
    tasks:    answer.tasks ?? answer.targets ?? [],
    deps:     answer.deps ?? [],
  }
}
