import type { DiscoverResult } from '../protocol/index.ts'

/**
 * Records the toolchain on every discovered project. The probe (`node
 * --version` and the like) runs only when there is a project, and once.
 * Returns the result it was given.
 */
export function setToolchain (result: DiscoverResult, probe: () => string): DiscoverResult {
  if (result.projects.length === 0) return result
  const toolchain = probe()
  for (const project of result.projects) project.toolchain = toolchain

  return result
}
