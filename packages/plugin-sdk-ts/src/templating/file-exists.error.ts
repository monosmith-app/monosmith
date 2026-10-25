/** A target file already exists: templates are written for a new project and never overwrite what is there. */
export class FileExistsError extends Error {
  /** The existing file, relative to the root. */
  readonly path: string

  constructor (path: string) {
    super(`${path}: file already exists`)
    this.name = 'FileExistsError'
    this.path = path
  }
}
