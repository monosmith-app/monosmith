import { readdirSync } from 'node:fs'
import { join, posix } from 'node:path'
import { isManifest } from '../projects/index.ts'

const SKIPPED_FOLDERS = new Set(['node_modules', '__pycache__'])

/**
 * The manifests of a plugin below a workspace, workspace-relative with forward
 * slashes: files named like `names` (see isManifest), skipping dot-folders and
 * dependency folders, and files at the root itself, as the Go conformance
 * suite does.
 */
export function findManifests (workspace: string, names: readonly string[]): string[] {
  const found: string[] = []
  const walk = (relative: string): void => {
    const folder = join(workspace, ...relative.split('/').filter(part => part !== ''))
    const entries = readdirSync(folder, { withFileTypes: true })
    for (const entry of entries) {
      const path = relative === '' ? entry.name : `${relative}/${entry.name}`
      if (entry.isDirectory()) {
        if (!entry.name.startsWith('.') && !SKIPPED_FOLDERS.has(entry.name)) walk(path)
      } else if (isManifest(names, entry.name) && posix.dirname(path) !== '.') {
        found.push(path)
      }
    }
  }
  try {
    walk('')
  } catch {
    // An unreadable folder ends the walk; what was found stays.
  }

  return found
}
