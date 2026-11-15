import { existsSync, readdirSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'

import { compareVersions } from '../version-choice/index.mjs'

/**
 * The folder that holds every smith the launchers and `smith update` installed:
 * `<home>/versions/<version>/smith`. SMITH_HOME moves it; default ~/.smith.
 *
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @returns {string} The versions folder.
 */
export function versionsDir (env) {
  return join(env.SMITH_HOME || join(homedir(), '.smith'), 'versions')
}

/**
 * The file name of the smith program on this system.
 *
 * @param {string} platform - process.platform.
 * @returns {string} smith.exe on Windows, smith elsewhere.
 */
export function programName (platform) {
  return platform === 'win32' ? 'smith.exe' : 'smith'
}

/**
 * Where one version's program lives.
 *
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @param {string} version - An exact version.
 * @param {string} platform - process.platform.
 * @returns {string} The program's path, whether it exists or not.
 */
export function programPath (env, version, platform) {
  return join(versionsDir(env), version, programName(platform))
}

/**
 * The newest installed version that has its program.
 *
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @param {string} platform - process.platform.
 * @returns {string | null} The version, or null when none is installed.
 */
export function newestInstalled (env, platform) {
  const dir = versionsDir(env)
  if (!existsSync(dir)) return null
  const installed = readdirSync(dir).filter(v => /^\d+\.\d+\.\d+/.test(v) && existsSync(programPath(env, v, platform)))

  return installed.toSorted(compareVersions).at(-1) ?? null
}
