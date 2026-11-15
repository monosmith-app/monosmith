import type { BlankKind } from './blank-folder.contract.ts'

/**
 * Entries that do not make a folder non-empty: the repository, the editor's
 * own settings and the files an operating system adds by itself. The Go side
 * (`smith new --here`) ignores the same names.
 */
const IGNORED = new Set(['.git', '.vscode', '.DS_Store', 'Thumbs.db', 'desktop.ini'])

/**
 * Tells whether a folder is empty from the names at its top level.
 * @param entries - the names in the folder, not their contents
 */
export function classifyFolder (entries: readonly string[]): BlankKind {
  return entries.every((entry) => IGNORED.has(entry)) ? 'empty' : 'unrecognised'
}
