import { posix, win32 } from 'node:path'
import type { RepositoryKind } from './repository-offer.contract.ts'

/** What the detection may look at; both are cheap reads of the folder's top level. */
export interface RepositoryProbe {
  exists: (path: string) => boolean
  /** The text of a file, or undefined when it cannot be read. */
  read:   (path: string) => string | undefined
}

/** The files that name a repository's tool, most specific first. */
const MARKERS: Array<{ file: string, kind: RepositoryKind }> = [
  { file: 'nx.json', kind: 'nx' },
  { file: 'turbo.json', kind: 'turborepo' },
  { file: 'lerna.json', kind: 'repository' },
  { file: 'pnpm-workspace.yaml', kind: 'repository' },
  { file: 'rush.json', kind: 'repository' },
  { file: 'go.work', kind: 'repository' },
]

/**
 * Tells whether a folder is a repository smith can bring in, from the files
 * at its top level: Nx and Turborepo by their config, other monorepo tools by
 * theirs, and a `package.json` that lists workspaces. Reads at most one small
 * file and never looks inside the folders.
 * @returns the kind and the file that showed it, or undefined for any other folder
 */
export function detectRepository (root: string, { exists, read }: RepositoryProbe): { kind: RepositoryKind, marker: string } | undefined {
  const path = root.includes('\\') ? win32 : posix
  const join = (...parts: string[]): string => path.join(...parts)
  for (const { file, kind } of MARKERS) {
    if (exists(join(root, file))) return { kind, marker: file }
  }
  const manifest = read(join(root, 'package.json'))

  return manifest !== undefined && listsWorkspaces(manifest) ? { kind: 'repository', marker: 'package.json' } : undefined
}

function listsWorkspaces (manifest: string): boolean {
  try {
    const parsed: unknown = JSON.parse(manifest)
    const workspaces = typeof parsed === 'object' && parsed !== null ? (parsed as { workspaces?: unknown }).workspaces : undefined

    return Array.isArray(workspaces) ? workspaces.length > 0 : typeof workspaces === 'object' && workspaces !== null
  } catch {
    return false
  }
}
