import { chmodSync, mkdirSync, renameSync, rmSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

import { checkArchive } from '../signature-check/index.mjs'
import { programName, programPath, versionsDir } from '../version-store/index.mjs'
import { readEntry } from '../zip-reading/index.mjs'
import { archiveName } from './archive-name.mapper.mjs'
import { releaseFile } from './releases.client.mjs'

/**
 * Downloads one smith version, checks it against the signed checksums.txt, and puts its program
 * in the versions folder. Nothing is written unless every check passes, and the program appears
 * in one rename, so a run that stops halfway leaves no half-written smith behind.
 *
 * @param {NodeJS.ProcessEnv} env - The environment.
 * @param {string} version - An exact version.
 * @param {string} platform - process.platform.
 * @param {string} arch - process.arch.
 * @param {(line: string) => void} say - Where progress goes (stderr).
 * @param {string} [publicKey] - The key the checksums must be signed with; default smith's release key.
 * @returns {Promise<string>} The installed program's path.
 * @throws {Error} When a file cannot be downloaded or does not pass the checks.
 */
export async function installVersion (env, version, platform, arch, say, publicKey) {
  const name = archiveName(version, platform, arch)
  say(`downloading smith ${version} (${name})`)
  const [archive, checksums, signature] = await Promise.all([
    releaseFile(env, version, name),
    releaseFile(env, version, 'checksums.txt'),
    releaseFile(env, version, 'checksums.txt.minisig'),
  ])
  const problem = checkArchive({ archive, name, checksums, signature: signature.toString('utf8') }, publicKey)
  if (problem) throw new Error(`smith ${version} was not installed: ${problem}`)
  const program = programName(platform)
  const bytes = readEntry(archive, entry => entry === program || entry.endsWith(`/${program}`))
  if (!bytes) throw new Error(`${name} holds no ${program}`)
  const target = programPath(env, version, platform)
  const staging = join(versionsDir(env), `.${version}-${process.pid}`)
  rmSync(staging, { recursive: true, force: true })
  mkdirSync(staging, { recursive: true })
  writeFileSync(join(staging, program), bytes)
  chmodSync(join(staging, program), 0o755)
  rmSync(dirname(target), { recursive: true, force: true })
  renameSync(staging, dirname(target))

  return target
}
