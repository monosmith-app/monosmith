import { posix, win32 } from 'node:path'

/** A folder of the editor's workspace. */
export interface FolderRef {
  /** Its path on disk. */
  root: string
  name: string
}

/**
 * The folders that are MonoSmith workspaces, which are the ones with a
 * `smith.toml` at their root. A multi-root workspace gets one session each.
 */
export function findSmithRoots (folders: readonly FolderRef[], exists: (path: string) => boolean): FolderRef[] {
  return folders.filter((folder) => {
    const manifest = folder.root.includes('\\') ? win32.join(folder.root, 'smith.toml') : posix.join(folder.root, 'smith.toml')

    return exists(manifest)
  })
}
