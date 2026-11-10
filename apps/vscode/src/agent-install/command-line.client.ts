import { execFile } from 'node:child_process'
import type { SmithRun } from '../smith-run/index.ts'

/** A `.cmd` file only runs through a shell on Windows; its path is then quoted. */
function launch (executable: string): { file: string, shell: boolean } {
  const shell = executable.toLowerCase().endsWith('.cmd')

  return { file: shell ? `"${executable}"` : executable, shell }
}

/**
 * Runs smith to its end in a folder and keeps what it printed. It never
 * rejects: a smith that cannot start is a run with exit code 1.
 * @param executable - the smith the folder runs
 * @param args - plain words, checked by the caller
 * @param cwd - the folder
 * @param timeoutMs - how long smith may run: one minute by default, longer for `smith add`, which installs and tests
 */
export async function runSmith (executable: string, args: readonly string[], cwd: string, timeoutMs = 60_000): Promise<SmithRun> {
  const { file, shell } = launch(executable)

  return await new Promise((resolve) => {
    execFile(file, [...args], { cwd, timeout: timeoutMs, maxBuffer: 16 * 1024 * 1024, windowsHide: true, shell }, (error, stdout, stderr) => {
      const exit = (error as { code?: unknown } | null)?.code
      const code = error === null ? 0 : (typeof exit === 'number' ? exit : 1)
      // A smith that did not start has no stderr; the error says why.
      resolve({ code, stdout, stderr: stderr === '' && typeof exit !== 'number' ? (error?.message ?? '') : stderr })
    })
  })
}

/** Whether the Claude Code command answers on the PATH of the editor. */
export async function hasClaudeCode (): Promise<boolean> {
  const shell = process.platform === 'win32'

  return await new Promise((resolve) => {
    execFile('claude', ['--version'], { timeout: 8000, windowsHide: true, shell }, (error) => { resolve(error === null) })
  })
}
