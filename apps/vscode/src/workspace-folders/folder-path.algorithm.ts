import { isAbsolute, relative, sep } from 'node:path'

/**
 * The path of a file relative to a workspace folder, with forward slashes, as smith words paths.
 * @param root - the workspace folder
 * @param file - an absolute path
 * @returns the relative path, or undefined when the file is outside the folder
 */
export function toFolderPath (root: string, file: string): string | undefined {
  const path = relative(root, file)

  return path === '' || path.startsWith('..') || isAbsolute(path) ? undefined : path.split(sep).join('/')
}

/**
 * The workspace folder that holds a file: the longest root that contains it.
 * @param roots - the folders that run smith
 * @param file - an absolute path
 */
export function folderOfFile (roots: readonly string[], file: string): string | undefined {
  return [...roots]
    .sort((a, b) => b.length - a.length)
    .find((root) => toFolderPath(root, file) !== undefined)
}
