import { execFile } from 'node:child_process'
import { parseVersion } from './parse-version.algorithm.ts'

/**
 * Asks an executable for its version with `--version`.
 * @returns the version, or undefined when it did not answer in time or did not say one
 */
export async function readSmithVersion (executable: string): Promise<string | undefined> {
  const shell = executable.toLowerCase().endsWith('.cmd')

  return await new Promise((resolve) => {
    execFile(
      shell ? `"${executable}"` : executable,
      ['--version'],
      { timeout: 5000, windowsHide: true, shell },
      (error, stdout) => { resolve(error === null ? parseVersion(stdout) : undefined) },
    )
  })
}
