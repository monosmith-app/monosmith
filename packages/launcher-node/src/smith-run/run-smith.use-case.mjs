import { spawnSync } from 'node:child_process'
import { existsSync } from 'node:fs'

import { installVersion, newestRelease } from '../release-download/index.mjs'
import { chooseVersion } from '../version-choice/index.mjs'
import { newestInstalled, programPath } from '../version-store/index.mjs'

/**
 * The smith program a run starts. A pinned version (SMITH_VERSION or the workspace's pin) is
 * downloaded when it is missing. Without a pin, the newest installed version runs, and only a
 * machine with none asks for the newest release: from then on `smith update` moves it.
 *
 * @param {object} host - What the run depends on.
 * @param {NodeJS.ProcessEnv} host.env - The environment.
 * @param {string} host.cwd - The current directory.
 * @param {string} host.platform - process.platform.
 * @param {string} host.arch - process.arch.
 * @param {(line: string) => void} host.say - Where progress goes.
 * @param {string} [host.publicKey] - The release key; default smith's (tests sign with their own).
 * @returns {Promise<string>} The program's path.
 */
export async function resolveProgram ({ env, cwd, platform, arch, say, publicKey }) {
  const pinned = chooseVersion(env, cwd)
  if (pinned) {
    const program = programPath(env, pinned, platform)

    return existsSync(program) ? program : installVersion(env, pinned, platform, arch, say, publicKey)
  }
  const installed = newestInstalled(env, platform)

  return installed ? programPath(env, installed, platform) : installVersion(env, await newestRelease(env), platform, arch, say, publicKey)
}

/**
 * Runs smith with the launcher's arguments and gives back its exit code. SMITH_LAUNCHER tells
 * smith it was started by this launcher, so `smith update` installs into the versions folder.
 *
 * @param {string} program - The smith program.
 * @param {string[]} args - The arguments.
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @param {string} launcher - The launcher's name and version (npm@3.0.0).
 * @returns {number} smith's exit code (1 when it could not start).
 */
export function runSmith (program, args, env, launcher) {
  const res = spawnSync(program, args, { stdio: 'inherit', env: { ...env, SMITH_LAUNCHER: launcher }, windowsHide: false })
  if (res.error) {
    process.stderr.write(`smith could not start: ${res.error.message}\n`)

    return 1
  }

  return res.status ?? 1
}
