import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

/** A version smith publishes: 2.50.0, 3.0.0-beta.1. */
const VERSION = /^\d+\.\d+\.\d+(?:-[0-9A-Z.-]+)?$/i

/**
 * The pin of the workspace a directory is in: `smith = "<version>"` under `[workspace]` in the
 * nearest smith.toml above it.
 *
 * @param {string} from - The directory to start from (the current directory).
 * @returns {string | null} The pinned version, or null when no smith.toml pins one (or pins latest).
 */
export function workspacePin (from) {
  for (let dir = from; ; dir = dirname(dir)) {
    const file = join(dir, 'smith.toml')
    if (existsSync(file)) return pinIn(readFileSync(file, 'utf8'))
    if (dirname(dir) === dir) return null
  }
}

/**
 * The `[workspace] smith` pin in smith.toml text.
 *
 * @param {string} text - The text of smith.toml.
 * @returns {string | null} The version, or null for none or `latest`.
 */
export function pinIn (text) {
  let table = ''
  for (const raw of text.split(/\r?\n/)) {
    const line = raw.replace(/#.*$/, '').trim()
    const header = /^\[([^\]]+)\]$/.exec(line)
    if (header) {
      table = header[1].trim()
      continue
    }
    const pin = /^smith\s*=\s*["']([^"']+)["']$/.exec(line)
    if (table === 'workspace' && pin) return VERSION.test(pin[1]) ? pin[1] : null
  }

  return null
}

/**
 * The version a run uses: SMITH_VERSION, else the workspace's pin, else none (the newest).
 *
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @param {string} cwd - The current directory.
 * @returns {string | null} An exact version, or null for the newest.
 * @throws {Error} When SMITH_VERSION is set to something that is not a version.
 */
export function chooseVersion (env, cwd) {
  const wanted = env.SMITH_VERSION?.trim()
  if (wanted && wanted !== 'latest') {
    if (!VERSION.test(wanted)) throw new Error(`SMITH_VERSION is "${wanted}": use a version such as 2.50.0, or latest`)

    return wanted
  }

  return wanted === 'latest' ? null : workspacePin(cwd)
}

/**
 * Compares two versions by their numbers; a prerelease sorts before its release.
 *
 * @param {string} a - A version.
 * @param {string} b - Another version.
 * @returns {number} Negative when a is older, positive when newer, 0 when equal.
 */
export function compareVersions (a, b) {
  const [mainA, preA = ''] = a.split('-', 2)
  const [mainB, preB = ''] = b.split('-', 2)
  const numsA = mainA.split('.').map(Number)
  const numsB = mainB.split('.').map(Number)
  for (let i = 0; i < 3; i++) if (numsA[i] !== numsB[i]) return numsA[i] - numsB[i]
  if (preA === preB) return 0
  if (preA === '') return 1
  if (preB === '') return -1

  return preA < preB ? -1 : 1
}
