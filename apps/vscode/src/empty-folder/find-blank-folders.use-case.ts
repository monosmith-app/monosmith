import type { FolderRef } from '../workspace-folders/index.ts'
import type { BlankFolder } from './blank-folder.contract.ts'
import { classifyFolder } from './classify-folder.policy.ts'

/**
 * The folders to offer a new workspace or a conversion in: the ones that
 * neither run smith nor look like a repository it can convert (`skip`). It
 * lists each folder's top level only and reads no file.
 * @param list - the names in a folder, or undefined when it cannot be listed
 */
export function findBlankFolders (folders: readonly FolderRef[], skip: ReadonlySet<string>, list: (root: string) => string[] | undefined): BlankFolder[] {
  const found: BlankFolder[] = []
  for (const folder of folders) {
    if (skip.has(folder.root)) continue
    const entries = list(folder.root)
    if (entries !== undefined) found.push({ ...folder, kind: classifyFolder(entries) })
  }

  return found
}
