import { posix, win32 } from 'node:path'

/** Where the smith that was found came from. */
export type SmithSource = 'setting' | 'workspace' | 'path'

/** What the search for smith needs to know. */
export interface SmithSearch {
  /** The `monosmith.path` setting; empty when not set. */
  setting:       string
  workspaceRoot: string
  /** Whether the user trusts this folder; an untrusted folder's own smith is never run. */
  trusted:       boolean
  platform:      NodeJS.Platform
  /** The `PATH` variable. */
  pathVariable:  string
  exists:        (path: string) => boolean
}

/** The smith to run and why it was chosen. */
export interface SmithLocation {
  executable: string
  source:     SmithSource
}

/** The answer of the search. */
export interface SmithLookup {
  /** The smith to run, or undefined when there is none. */
  location?:        SmithLocation
  /** The first smith on the `PATH`, whichever won; what a terminal would run. */
  pathSmith?:       string
  /** The folder holds its own smith, which was left alone because the folder is not trusted. */
  skippedUntrusted: boolean
}

/**
 * Finds the smith the team has installed, never one the extension carries:
 * the setting, then the workspace's own install (`node_modules/.bin`,
 * `.smith/bin`) when the folder is trusted, then the `PATH`. The first that
 * exists wins.
 */
export function locateSmith (search: SmithSearch): SmithLookup {
  const windows = search.platform === 'win32'
  const path = windows ? win32 : posix
  const own = [
    path.join(search.workspaceRoot, 'node_modules', '.bin', windows ? 'smith.cmd' : 'smith'),
    path.join(search.workspaceRoot, '.smith', 'bin', windows ? 'smith.exe' : 'smith'),
  ].filter((candidate) => search.exists(candidate))
  const names = windows ? ['smith.exe', 'smith.cmd'] : ['smith']
  const pathSmith = search.pathVariable
    .split(path.delimiter)
    .filter((dir) => dir !== '')
    .flatMap((dir) => names.map((name) => path.join(dir, name)))
    .find((candidate) => search.exists(candidate))
  const skippedUntrusted = !search.trusted && own.length > 0
  const location = pick(search.setting.trim(), search.trusted ? own[0] : undefined, pathSmith)

  return { location, pathSmith, skippedUntrusted }
}

function pick (setting: string, workspace: string | undefined, onPath: string | undefined): SmithLocation | undefined {
  if (setting !== '') return { executable: setting, source: 'setting' }
  if (workspace !== undefined) return { executable: workspace, source: 'workspace' }

  return onPath === undefined ? undefined : { executable: onPath, source: 'path' }
}
