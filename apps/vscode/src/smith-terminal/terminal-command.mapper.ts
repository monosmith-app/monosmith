/** The families of shell whose quoting differs. */
export type ShellFamily = 'powershell' | 'cmd' | 'posix'

/** Tells which family of shell a terminal runs, from the path of its executable. */
export function shellFamily (shell: string): ShellFamily {
  const name = shell.split(/[\\/]/).at(-1)?.toLowerCase().replace(/\.exe$/, '') ?? ''
  if (name === 'powershell' || name === 'pwsh') return 'powershell'

  return name === 'cmd' ? 'cmd' : 'posix'
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
