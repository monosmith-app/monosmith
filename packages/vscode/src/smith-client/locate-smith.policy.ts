import { posix, win32 } from 'node:path'

/** What the search for smith needs to know. */
export interface SmithSearch {
  /** The `monosmith.path` setting; empty when not set. */
  setting:       string
  workspaceRoot: string
  platform:      NodeJS.Platform
  /** The `PATH` variable. */
  pathVariable:  string
  exists:        (path: string) => boolean
}

/**
 * Finds the smith the team has installed, never one the extension carries:
 * the setting, then the workspace's own install (`node_modules/.bin`,
 * `.smith/bin`), then the `PATH`. The first that exists wins.
 * @returns the executable, or undefined when there is none
 */
export function locateSmith (search: SmithSearch): string | undefined {
  const windows = search.platform === 'win32'
  if (search.setting.trim() !== '') return search.setting.trim()
  const path = windows ? win32 : posix
  const own = [
    path.join(search.workspaceRoot, 'node_modules', '.bin', windows ? 'smith.cmd' : 'smith'),
    path.join(search.workspaceRoot, '.smith', 'bin', windows ? 'smith.exe' : 'smith'),
  ]
  const names = windows ? ['smith.exe', 'smith.cmd'] : ['smith']
  const onPath = search.pathVariable
    .split(path.delimiter)
    .filter((dir) => dir !== '')
    .flatMap((dir) => names.map((name) => path.join(dir, name)))

  return [...own, ...onPath].find((candidate) => search.exists(candidate))
}
