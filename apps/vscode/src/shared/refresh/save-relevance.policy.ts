/** Folders whose files a save in does not change what a run would find. */
const IGNORED = new Set(['node_modules', '.git', 'dist', 'out', '.smith', '.nx', '.turbo'])

/**
 * Whether saving a file can change which projects are affected: it is inside
 * the folder and not in a folder that holds installs, builds or smith's own
 * state.
 */
export function isRelevantSave (root: string, file: string): boolean {
  const base = normalise(root)
  const target = normalise(file)
  const windows = /^[a-z]:/i.test(base)
  const inside = windows ? target.toLowerCase().startsWith(`${base.toLowerCase()}/`) : target.startsWith(`${base}/`)

  return inside && target.slice(base.length + 1).split('/').every((part) => !IGNORED.has(part))
}

function normalise (path: string): string {
  return path.replaceAll('\\', '/').replace(/\/+$/, '')
}
