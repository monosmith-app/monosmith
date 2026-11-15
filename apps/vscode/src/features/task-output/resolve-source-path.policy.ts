import { isAbsolute, join, normalize } from 'node:path'

/** What is known about a path a tool printed. */
export interface PathSearch {
  /** The path as printed. */
  file:   string
  /**
   * The directories a relative path may be relative to, the likeliest first:
   * the directory the task ran in (its project), then the workspace folder.
   */
  bases:  readonly string[]
  exists: (path: string) => boolean
}

/**
 * Finds the file a tool meant. Tools print paths relative to where they ran,
 * which is the project's directory and not the workspace folder, so a relative
 * path is tried against each base in turn.
 * @returns the path of a file that exists, or undefined
 */
export function resolveSourcePath (search: PathSearch): string | undefined {
  if (isAbsolute(search.file)) return search.exists(search.file) ? normalize(search.file) : undefined

  return search.bases.map((base) => join(base, search.file)).find((candidate) => search.exists(candidate))
}
