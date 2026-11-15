/** Node's names for systems and processors, as smith's archives spell them. */
const SYSTEMS = { darwin: 'darwin', linux: 'linux', win32: 'windows' }
const PROCESSORS = { x64: 'amd64', arm64: 'arm64' }

/**
 * The name of the release archive for one version on this machine:
 * monosmith_<version>_<os>_<arch>.zip.
 *
 * @param {string} version - An exact version.
 * @param {string} platform - process.platform.
 * @param {string} arch - process.arch.
 * @returns {string} The archive's file name.
 * @throws {Error} For a system or processor smith is not built for.
 */
export function archiveName (version, platform, arch) {
  const system = SYSTEMS[platform]
  const processor = PROCESSORS[arch]
  if (!system || !processor) throw new Error(`smith is not built for ${platform} on ${arch}: it runs on macOS, Linux and Windows, on x64 and arm64`)

  return `monosmith_${version}_${system}_${processor}.zip`
}
