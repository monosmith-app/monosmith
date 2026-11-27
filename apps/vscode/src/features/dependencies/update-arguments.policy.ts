const PACKAGE_NAME = /^[\w@][\w./@-]*$/

/**
 * The arguments that update packages with `smith deps up` (the command the terminal interface
 * runs), or undefined when a name could be read as an option or is not a package name at all.
 * A name is one argument, never a flag. A major release takes `--major`.
 * @param names - one package, or several marked together
 * @param major - true when one of them moves to a new major version
 * @returns the arguments, or undefined
 */
export function updateArguments (names: string | readonly string[], major = false): string[] | undefined {
  const list = typeof names === 'string' ? [names] : [...names]
  if (list.length === 0 || list.some((name) => !PACKAGE_NAME.test(name))) return undefined

  return ['deps', 'up', ...list, ...(major ? ['--major'] : [])]
}

/**
 * The arguments that let a person pick the updates in the terminal.
 * @returns `deps up`
 */
export function pickArguments (): string[] {
  return ['deps', 'up']
}

/**
 * The arguments that open the dependency repair in the terminal: smith asks what to repair
 * and shows the plan before it changes anything.
 * @returns `deps repair`
 */
export function repairArguments (): string[] {
  return ['deps', 'repair']
}
