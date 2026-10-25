import type { Project } from '../protocol/index.ts'

/**
 * Whether a release publishes the project to a registry: a package, unless
 * `smith.toml` says `registry = "none"`. Plugins list the credentials a
 * release needs only for such projects.
 */
export function publishes (project: Pick<Project, 'kind' | 'registries'>): boolean {
  return project.kind === 'package' && project.registries?.noPublish !== true
}
