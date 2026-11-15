/** The families of shell whose quoting differs. */
export type ShellFamily = 'powershell' | 'cmd' | 'posix'

/** Tells which family of shell a terminal runs, from the path of its executable. */
export function shellFamily (shell: string): ShellFamily {
  const name = shell.split(/[\\/]/).at(-1)?.toLowerCase().replace(/\.exe$/, '') ?? ''
  if (name === 'powershell' || name === 'pwsh') return 'powershell'

  return name === 'cmd' ? 'cmd' : 'posix'
}

/** The line to type in a terminal, and the folder that must be on its PATH for the line to run. */
export interface TerminalLaunch {
  line:       string
  /** The folder of smith to put first on the PATH of the terminal; absent when the line is the full path. */
  directory?: string
}

/**
 * The launch of smith in a terminal. When the executable is called `smith`, the
 * line is `smith <arguments>` and its folder goes first on the PATH of the terminal,
 * so the person reads the command they could type themselves. Any other
 * name keeps the full, quoted path.
 * @param executable - the smith the folder runs
 * @param args - plain words, checked by the caller
 * @param shell - the shell of the terminal
 * @returns the line and the folder to put on the PATH
 */
export function terminalLaunch (executable: string, args: readonly string[], shell: string): TerminalLaunch {
  const cut = Math.max(executable.lastIndexOf('/'), executable.lastIndexOf('\\'))
  const directory = cut > 0 ? executable.slice(0, cut) : ''
  const base = executable.slice(cut + 1).toLowerCase().replace(/\.(?:exe|cmd|bat|ps1)$/, '')

  return base !== 'smith' || directory === '' ? { line: toTerminalCommand(executable, args, shell) } : { line: `smith ${args.join(' ')}`, directory }
}

/**
 * The line to type in a terminal to run smith with the arguments, with the
 * path of the executable quoted for the shell. The arguments are plain words
 * (the caller checked them), so they are not quoted.
 */
export function toTerminalCommand (executable: string, args: readonly string[], shell: string): string {
  const family = shellFamily(shell)
  const rest = args.join(' ')
  switch (family) {
    case 'powershell': {
      return `& '${executable.replaceAll("'", "''")}' ${rest}`
    }
    case 'cmd': {
      return `"${executable}" ${rest}`
    }
    case 'posix': {
      return `'${executable.replaceAll("'", String.raw`'\''`)}' ${rest}`
    }
  }
}
