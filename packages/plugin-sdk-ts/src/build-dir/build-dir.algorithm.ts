/**
 * The directory at the workspace root every build writes into: smith's opinion,
 * so all output is in one place, outside every source tree, and cleaning is
 * deleting one folder. A plugin builds a project into {@link projectBuildDir};
 * `smith doctor` fails a task whose outputs fall outside it.
 */
export const BUILD_DIR = 'dist'

function trimSlashes (root: string): string {
  return root.replaceAll(/^\/+|\/+$/g, '')
}

/** Where a project's build goes, workspace-relative: `dist/<project dir>`, mirroring the workspace (`dist/apps/web`). */
export function projectBuildDir (root: string): string {
  return `${BUILD_DIR}/${trimSlashes(root)}`
}

/**
 * The same directory relative to the project (`../../dist/apps/web`): what a
 * tool configured from inside the project, or a task's `outputs`, name.
 */
export function buildDirFromProject (root: string): string {
  const trimmed = trimSlashes(root)

  return '../'.repeat(trimmed.split('/').length) + projectBuildDir(trimmed)
}

/** Whether a workspace-relative path is inside the build directory. */
export function inBuildDir (path: string): boolean {
  return path === BUILD_DIR || path.startsWith(`${BUILD_DIR}/`)
}
