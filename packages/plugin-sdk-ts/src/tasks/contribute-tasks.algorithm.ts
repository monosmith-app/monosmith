import type { ContributeResult, Contribution, Project } from '../protocol/index.ts'

/**
 * The answer to `contribute` when every project gets the same kind of
 * addition: calls `contribution` for each project and keys the result by
 * project name. Return `undefined` for a project the plugin has nothing to add
 * to; it is left out.
 *
 * ```ts
 * contribute: ({ projects }) => contributeTo(projects, () => ({
 *   tasks: { 'layout-check': task(['node', checker, 'layout-check'], { cache: true }) },
 *   dependsOn: { lint: ['layout-check'] },
 * }))
 * ```
 */
export function contributeTo (projects: readonly Project[], contribution: (project: Project) => Contribution | undefined): ContributeResult {
  const result: Record<string, Contribution> = {}
  for (const project of projects) {
    const added = contribution(project)
    if (added !== undefined) result[project.name] = added
  }

  return { projects: result }
}
